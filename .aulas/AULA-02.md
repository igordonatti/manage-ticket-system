# Aula 2 — Redis como estado operacional

Nesta aula o Redis deixa de ser apenas um container ligado e passa a fazer parte do comportamento observável da API. A entrega completa será dividida em três checkpoints.

## Resultado esperado da aula

Ao final:

- a API mantém um único cliente Redis durante sua execução;
- `/health` informa se o Redis está disponível;
- eventos e sua disponibilidade sobrevivem ao reinício da API;
- falhas do Redis viram respostas HTTP controladas;
- o código de domínio continua sem importar `go-redis`.

## Ideia central

Neste projeto, Redis não é uma cópia descartável do PostgreSQL. Ele contém o estado operacional usado no caminho crítico de venda. Se esse estado estiver indisponível, a aplicação não pode fingir que está saudável nem consultar silenciosamente outro banco.

A direção das dependências será:

```text
cmd/api              configura e conecta as partes
   │
   ├── httpapi       conhece somente um contrato de verificação
   ├── event         contém regras de domínio
   └── redisstore    conhece o cliente go-redis
                         │
                         └── Redis
```

## Checkpoint 2A — conhecer o Redis pelo CLI

Antes de escrever Go, execute cada comando separadamente e observe a resposta:

```bash
docker compose exec -T redis redis-cli PING

docker compose exec -T redis redis-cli SET study:aula2:greeting ola EX 60
docker compose exec -T redis redis-cli GET study:aula2:greeting
docker compose exec -T redis redis-cli TTL study:aula2:greeting

docker compose exec -T redis redis-cli HSET study:aula2:event:demo name "Evento Demo" capacity 100 price_cents 5000
docker compose exec -T redis redis-cli TYPE study:aula2:event:demo
docker compose exec -T redis redis-cli HGET study:aula2:event:demo name
docker compose exec -T redis redis-cli HGETALL study:aula2:event:demo

docker compose exec -T redis redis-cli DEL study:aula2:event:demo study:aula2:greeting
```

Observe:

- uma chave Redis possui um tipo;
- uma string armazena um valor diretamente;
- um hash agrupa pares campo/valor dentro de uma chave;
- Redis devolve números como representações textuais ao cliente, que precisa convertê-los;
- TTL é associado à chave, não automaticamente à ideia de entidade da aplicação.

### Perguntas do checkpoint 2A

1. Qual é a diferença entre a chave `study:aula2:event:demo` e o campo `name` dentro dela?
2. O que `TTL` retorna enquanto a chave expira? Pesquise também o significado de `-1` e `-2`.
3. Por que `capacity` volta do Redis como texto mesmo tendo sido enviado como número?

## Checkpoint 2B — cliente Go e health check real

Implemente este checkpoint antes do repositório de eventos.

### Dependência

Dentro de `backend`:

```bash
go get github.com/redis/go-redis/v9
```

Use o cliente oficial `go-redis/v9`. Não escreva cliente RESP manualmente e não adicione outro framework.

### Arquivos envolvidos

```text
backend/
├── cmd/api/main.go
├── internal/httpapi/health.go
└── internal/redisstore/client.go
```

### Responsabilidade de `redisstore`

Crie um tipo que encapsule `*redis.Client` e ofereça somente o que a aplicação precisa neste checkpoint:

```go
Ping(ctx context.Context) error
Close() error
```

Requisitos:

- o endereço é recebido pelo construtor; `redisstore` não lê variável de ambiente;
- o cliente deve ser criado uma única vez no início do processo;
- configure timeouts de conexão, leitura e escrita;
- `Ping` deve propagar o contexto recebido;
- `Close` deve fechar o cliente real;
- não use variável global;
- não crie um cliente novo dentro de cada requisição.

### Configuração no `main`

O `main` é o ponto de composição. Ele deve:

1. ler `REDIS_ADDR`, usando `localhost:6379` como padrão local;
2. criar o cliente Redis;
3. garantir seu fechamento no fim do processo;
4. injetá-lo no health handler;
5. preservar o shutdown gracioso da Aula 0.

Criar `redis.Client` não comprova que o servidor está disponível. A verificação real acontece quando um comando, como `PING`, é executado.

### Contrato consumido pelo handler

O pacote `httpapi` não deve importar `go-redis`. Declare nele uma interface mínima, implementada implicitamente pelo seu adaptador:

```go
type RedisChecker interface {
	Ping(ctx context.Context) error
}
```

Transforme o health handler para receber essa dependência. Você pode usar uma struct que implementa `http.Handler` ou um construtor que devolve `http.Handler`; explique sua escolha na revisão.

### Comportamento HTTP

Redis disponível:

```http
HTTP/1.1 200 OK
Content-Type: application/json
```

```json
{
  "status": "ok",
  "redis": "ok",
  "version": "dev"
}
```

Redis indisponível ou fora do prazo:

```http
HTTP/1.1 503 Service Unavailable
Content-Type: application/json
```

```json
{
  "status": "degraded",
  "redis": "unavailable",
  "version": "dev"
}
```

Use um contexto derivado de `r.Context()` com um timeout curto para o `PING`. Não devolva ao cliente a mensagem interna de conexão, endereço ou senha.

Neste laboratório `/health` funciona como readiness. Num sistema real, normalmente separaríamos liveness, que verifica o processo, de readiness, que verifica suas dependências.

### Verificação manual

Com Redis disponível:

```bash
go run ./cmd/api
curl -i http://localhost:8080/health
```

Em seguida:

```bash
docker compose stop redis
curl -i http://localhost:8080/health
docker compose start redis
curl -i http://localhost:8080/health
```

O segundo `curl` deve retornar 503 sem travar até os timeouts gerais do servidor. O último deve voltar a retornar 200 sem reiniciar a API.

### Limites deste checkpoint

Ainda não implemente:

- criação ou listagem de eventos via HTTP;
- hashes de eventos;
- disponibilidade;
- retries automáticos;
- Redis Cluster;
- scripts Lua;
- variáveis globais.

Eu escreverei os testes do health handler usando um checker falso depois que você terminar a implementação.

### Perguntas do checkpoint 2B

1. Por que construir `redis.Client` não garante que existe uma conexão funcional?
2. Por que o contexto do `PING` deve nascer de `r.Context()` em vez de `context.Background()`?
3. O que ganhamos fazendo o handler depender da interface `RedisChecker` em vez de `*redis.Client`?
4. Por que o cliente Redis deve ser compartilhado pelo processo em vez de criado por requisição?
5. Qual é a diferença entre `DialTimeout`, `ReadTimeout` e o deadline do contexto?

## Checkpoint 2C — repositório de eventos e inventário

Neste checkpoint separaremos o contrato de persistência, pertencente à aplicação, de sua implementação com Redis. Criaremos:

```text
internal/event/repository.go
internal/redisstore/event_repository.go
internal/redisstore/event_repository_test.go
```

O trabalho será incremental:

1. declarar no domínio o contrato mínimo de criação e leitura;
2. definir a tradução entre `event.Event` e campos textuais do Redis;
3. criar o evento e seu inventário sem reinicializar disponibilidade existente;
4. reconstruir um evento a partir do hash;
5. indexar eventos pela data de início;
6. comprovar com Redis real que os dados sobrevivem à troca do cliente da aplicação.

Um evento será armazenado num hash. O inventário ficará em outro hash, pois no estilo baseado em espaço ele é estado operacional acessado e alterado com frequência. Um sorted set funcionará como índice ordenado pela data do evento:

```text
event:{event-001}       HASH
inventory:{event-001}   HASH
events:by_start         SORTED SET
```

Os campos de tempo serão armazenados em formato textual explícito e convertidos de volta para `time.Time`. Capacidade e preço também voltarão como texto e precisarão de conversão segura antes da reconstrução do domínio.

As chaves com `{event-001}` preparam afinidade entre evento e inventário para Redis Cluster: o conteúdo entre chaves é a hash tag usada na escolha do slot. O índice global `events:by_start`, porém, não pertence ao mesmo slot. Isso significa que uma transação envolvendo as três chaves funciona no Redis único deste laboratório, mas exigirá outra estratégia quando distribuirmos os dados entre vários nós. Não chamaremos essa implementação de pronta para Cluster antes de resolver esse limite.

### Critérios de aceite do checkpoint 2C

- o pacote `event` não importa `go-redis`;
- o adaptador Redis implementa implicitamente o contrato do domínio;
- criar um evento produz o hash do evento, o inventário inicial e a entrada no índice;
- uma segunda criação com o mesmo ID não reinicializa o inventário;
- buscar um ID inexistente retorna um erro de domínio conhecido;
- dados inválidos ou corrompidos no Redis não geram um `Event` inválido;
- fechar um cliente, criar outro e buscar o mesmo evento preserva os dados;
- os testes de integração limpam somente as chaves que eles próprios criaram.

## Referências oficiais

- Cliente Go: https://redis.io/docs/latest/develop/clients/go/
- Conexão com Go: https://redis.io/docs/latest/develop/clients/go/connect/
- Hashes: https://redis.io/docs/latest/develop/data-types/hashes/
- Sorted sets: https://redis.io/docs/latest/develop/data-types/sorted-sets/
