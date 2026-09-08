# Aula 3 — reservas e idempotência

Nesta aula construiremos a primeira mutação do caminho crítico de venda: reservar ingressos por dez minutos. O objetivo principal é compreender idempotência e tornar visíveis os limites de uma sequência `read-modify-write` antes de introduzirmos atomicidade com Lua na Aula 4.

Esta aula só deve começar depois que todos os critérios do checkpoint 2C forem concluídos.

## Como trabalharemos

Cada checkpoint seguirá o mesmo ciclo das aulas anteriores:

1. revisar o conceito e o problema que o bloco resolve;
2. receber e escrever um bloco de código pequeno;
3. explicar as partes relevantes do código;
4. executar formatação e testes;
5. responder às perguntas do checkpoint;
6. receber revisão antes de avançar.

Durante a orientação pelo chat, os blocos necessários serão fornecidos completos e explicados continuamente. Os testes serão escritos ou complementados pelo agente conforme o acordo do plano de estudo; a implementação Go e Redis continuará sendo realizada pelo aluno.

## Revisão das aulas anteriores

Antes de criar reservas, precisamos recuperar os conceitos que já sustentam a aplicação.

### Aula 0 — processo HTTP e ciclo de vida

Construímos uma API com a biblioteca padrão do Go e aprendemos que:

- `http.ResponseWriter` representa a resposta que o handler está construindo;
- `*http.Request` contém método, URL, headers, corpo e o contexto da requisição;
- structs Go precisam ser serializadas para produzir JSON na resposta HTTP;
- timeouts do `http.Server` limitam leituras, escritas e conexões ociosas;
- `ListenAndServe` bloqueia e por isso executa numa goroutine enquanto `main` coordena sinais;
- `Shutdown` deixa de aceitar novas conexões e aguarda requisições em andamento dentro do prazo;
- `defer` agenda uma função para o fim da função que o declarou;
- `log.Fatal` e `os.Exit` não executam funções adiadas com `defer`.

Fluxo principal:

```text
main cria servidor
      ↓
ListenAndServe executa em uma goroutine
      ↓
main aguarda SIGINT, SIGTERM ou erro do servidor
      ↓
Shutdown recebe um contexto com prazo
      ↓
recursos são fechados
```

### Aula 1 — domínio e invariantes

Criamos `event.Event` sem dependência de HTTP ou Redis e aprendemos que:

- uma entidade protege regras de negócio independentemente da forma de entrada;
- campos não exportados impedem alteração direta fora do pacote;
- o construtor `event.New` rejeita estados inválidos;
- métodos de leitura expõem somente os dados necessários;
- erros sentinela permitem identificar regras específicas com `errors.Is`;
- preço é armazenado em centavos inteiros para evitar arredondamentos de ponto flutuante;
- datas fixas tornam testes determinísticos;
- a entidade não recebe tags JSON porque o domínio não é um DTO HTTP.

Invariantes principais do evento:

```text
ID não vazio
nome e local válidos
capacity > 0
priceCents >= 0
salesStart < eventStart
status inicial = draft
```

### Aula 2 — Redis como estado operacional

Transformamos Redis em uma dependência observável e começamos a persistir o estado operacional:

- uma chave Redis possui um tipo, como string, hash ou sorted set;
- hashes agrupam pares campo/valor, mas seus valores chegam ao cliente como texto/bytes;
- TTL pertence à chave: `-1` significa sem expiração e `-2` significa chave inexistente;
- `redis.Options` configura conexão, leitura, escrita e respeito ao contexto;
- criar `redis.Client` não comprova conectividade; um comando como `PING` realiza a verificação;
- um único cliente compartilhado reutiliza um pool de conexões;
- `RedisChecker` impede que o handler dependa diretamente de `go-redis`;
- `r.Context()` propaga cancelamento e prazo da requisição até o Redis;
- o `main` funciona como ponto de composição das implementações concretas;
- `event.Repository` descreve o que a aplicação precisa sem conhecer Redis;
- o adaptador `redisstore` traduz entidades para hashes e erros Redis para erros da aplicação;
- `event:{id}` e `inventory:{id}` usam a mesma hash tag para preparar afinidade por evento;
- `events:by_start` é um índice global e não pertence automaticamente ao mesmo slot das chaves do evento.

A separação construída até aqui é:

```text
cmd/api
   ├── cria redisstore.Client
   ├── cria redisstore.EventRepository
   └── injeta contratos nos handlers

httpapi ── depende de interfaces pequenas
event   ── contém entidades, invariantes e contratos
redisstore ── conhece go-redis e implementa os contratos
```

### Revisão ativa antes da Aula 3

Responda sem consultar o código primeiro; depois confira sua resposta:

1. Qual é a diferença entre o timeout do servidor, o timeout do cliente Redis e o deadline de uma operação?
2. Por que `event.Event` não possui tags JSON nem importa `go-redis`?
3. Por que o handler depende de uma interface, mas o `main` cria a implementação concreta?
4. O que aconteceria se um cliente Redis novo fosse criado em cada requisição?
5. Por que capacidade e preço precisam ser convertidos depois de um `HGETALL`?
6. Qual é a diferença entre a identidade `event-001`, a chave `event:{event-001}` e o campo `id` dentro do hash?
7. Por que evento e inventário são armazenados separadamente?
8. O que `go test -race ./...` procura que `go vet ./...` não procura?

## Resultado esperado da Aula 3

Ao final:

- uma reserva válida nasce com status `pending` e expiração de dez minutos;
- quantidade aceita fica entre 1 e 10 ingressos;
- `POST /events/{eventId}/reservations` exige `Idempotency-Key`;
- repetir sequencialmente a mesma requisição devolve a mesma reserva;
- repetir uma chave com conteúdo diferente produz conflito controlado;
- o inventário é reduzido durante a criação da reserva;
- reserva, idempotência e agenda de expiração são visíveis no Redis;
- falhas de domínio e infraestrutura viram respostas HTTP explícitas;
- um experimento demonstra por que essa primeira sequência ainda não é segura sob concorrência;
- a necessidade da operação atômica da Aula 4 fica comprovada, não apenas presumida.

## Modelo de chaves

As chaves relacionadas à mutação de um evento usarão sua identidade como hash tag:

```text
inventory:{event-001}                         HASH
reservation:{event-001}:reservation-001      HASH
idempotency:{event-001}:<chave>               HASH ou STRING
reservations:expiring:{event-001}             SORTED SET
```

O trecho `{event-001}` mantém as chaves da operação no mesmo slot quando chegarmos ao Redis Cluster.

Nunca coloque uma `Idempotency-Key` arbitrária diretamente numa chave Redis sem definir codificação ou validação. Ela pode conter espaços, separadores ou caracteres que alterem nossa convenção de nomes.

## Checkpoint 3A — domínio da reserva

Criaremos:

```text
internal/reservation/
├── reservation.go
├── repository.go
└── reservation_test.go
```

A entidade possuirá conceitualmente:

- ID da reserva;
- ID do evento;
- quantidade;
- status;
- instante de criação;
- instante de expiração.

Regras iniciais:

- IDs da reserva e do evento não podem ser vazios;
- quantidade deve estar entre 1 e 10;
- uma reserva nova começa como `pending`;
- expiração deve ocorrer depois da criação;
- a duração usada pelo caso de uso será de dez minutos;
- domínio não chama `time.Now()` diretamente;
- domínio não conhece TTL, Redis, HTTP ou `Idempotency-Key`.

O contrato de persistência oferecerá somente as operações necessárias para criar e recuperar uma reserva nesta aula.

### Perguntas do checkpoint 3A

1. Por que a entidade possui `ExpiresAt` se o Redis também pode possuir TTL?
2. Qual é a diferença entre a regra “expira depois da criação” e a política “dura dez minutos”?
3. Por que não devemos chamar `time.Now()` dentro de todos os métodos que precisam do horário atual?
4. O que campos não exportados impedem numa reserva?
5. Por que `pending` é um tipo de domínio em vez de uma string espalhada pelo código?

## Checkpoint 3B — idempotência pelo Redis CLI

Antes do Go, faremos um experimento com `SET`, `NX`, `GET`, hashes e TTL.

Precisaremos distinguir:

```text
mesma chave + mesmo conteúdo      → devolver resultado anterior
mesma chave + conteúdo diferente  → conflito de idempotência
chave nova                         → tentar nova reserva
```

A idempotência não significa “ignorar toda repetição”. Ela associa uma intenção externa a um resultado estável.

Também definiremos por quanto tempo o registro deve sobreviver. O TTL da idempotência não pode desaparecer antes do período no qual uma repetição legítima ainda pode chegar.

### Perguntas do checkpoint 3B

1. Qual problema o modificador `NX` resolve?
2. Por que armazenar somente o ID da reserva pode ser insuficiente para detectar uma chave reutilizada com outra quantidade?
3. Qual é a diferença entre TTL da chave de idempotência e `expires_at` da reserva?
4. O que pode acontecer se executarmos `GET` e `SET` como comandos independentes sob concorrência?
5. A idempotência elimina a necessidade de controle de estoque? Por quê?

## Checkpoint 3C — serviço de criação e baseline read-modify-write

Criaremos um serviço de aplicação que coordena domínio e persistência. Seu fluxo inicial será observável em passos separados:

```text
validar comando
      ↓
consultar idempotência
      ↓
ler disponibilidade
      ↓
verificar quantidade
      ↓
reduzir disponibilidade
      ↓
gravar reserva e expiração
      ↓
gravar resultado idempotente
```

Essa sequência é intencionalmente a baseline da Aula 3. Ela ajuda a enxergar duas classes diferentes de problema:

- concorrência: duas requisições podem tomar decisões usando o mesmo valor antigo;
- falha parcial: o processo pode parar depois de reduzir estoque e antes de registrar a idempotência.

Não esconderemos esses limites com mutex local. Um mutex protegeria somente uma instância da API e contrariaria a futura distribuição baseada em espaço.

### Perguntas do checkpoint 3C

1. O que significa `read-modify-write` neste fluxo?
2. Como duas requisições com chaves de idempotência diferentes podem deixar o estoque negativo?
3. Por que repetir automaticamente toda a sequência após um timeout pode consumir estoque duas vezes?
4. Por que um mutex Go não resolve o problema quando existem várias instâncias da API?
5. Qual falha pode ocorrer entre reduzir disponibilidade e gravar a reserva?
6. Por que todas as chaves da futura operação atômica devem compartilhar `{eventId}`?

## Checkpoint 3D — endpoint de reserva

Implementaremos:

```http
POST /events/{eventId}/reservations
Idempotency-Key: checkout-usuario-123
Content-Type: application/json
```

```json
{
  "quantity": 2
}
```

Comportamentos mínimos:

- primeira execução bem-sucedida: `201 Created`;
- repetição com a mesma chave e mesmo conteúdo: `200 OK` e a mesma reserva;
- chave ausente ou entrada inválida: `422 Unprocessable Entity`;
- evento inexistente: `404 Not Found`;
- estoque insuficiente: `409 Conflict`;
- chave reutilizada com outra intenção: `409 Conflict`;
- Redis indisponível: `503 Service Unavailable`;
- erros internos não revelam endereço, senha ou detalhes do Redis.

O handler será responsável por HTTP e JSON. O serviço será responsável pelo fluxo do caso de uso. A entidade continuará responsável por suas invariantes.

### Perguntas do checkpoint 3D

1. Por que a primeira resposta usa `201`, mas uma repetição idempotente usa `200`?
2. Por que `Idempotency-Key` fica no header em vez de ser o ID da entidade criado pelo domínio?
3. Qual camada deve converter `ErrInsufficientInventory` em HTTP 409?
4. Por que o handler não deve comparar erros do `go-redis` diretamente?
5. Como o contexto da requisição chega até os comandos Redis?

## Checkpoint 3E — experimento de concorrência

Executaremos requisições simultâneas contra um evento pequeno e observaremos a implementação baseline. O objetivo não é aprová-la como segura, mas produzir evidência do problema que a Aula 4 resolverá.

O experimento deverá registrar:

- capacidade inicial;
- quantidade de requisições concorrentes;
- chaves de idempotência usadas;
- reservas aceitas;
- estoque final;
- presença de estoque negativo, reservas sem idempotência ou outros estados parciais.

### Perguntas do checkpoint 3E

1. Qual invariável foi violada ou ficou vulnerável?
2. O problema apareceu entre goroutines da mesma API ou também apareceria entre processos diferentes?
3. Por que os testes unitários sequenciais não descobriram esse comportamento?
4. Qual conjunto de passos precisa virar uma única operação atômica?
5. O que esperamos que Lua ou Redis Functions mudem na Aula 4?

## Critérios de aceite

- `go fmt ./...` não produz alterações;
- `go test -race ./...` passa para o comportamento implementado;
- `go vet ./...` passa;
- testes unitários usam relógio e gerador de ID controláveis;
- repetição sequencial da mesma `Idempotency-Key` não reduz estoque novamente;
- reutilização da chave com outra quantidade é rejeitada;
- o endpoint não importa `go-redis`;
- o domínio não conhece Redis ou HTTP;
- as limitações concorrentes da baseline são documentadas com evidência;
- a implementação não é apresentada como atomicamente segura antes da Aula 4.

## Limites desta aula

Ainda não implementaremos:

- scripts Lua ou Redis Functions;
- garantia atômica entre estoque, reserva, expiração e idempotência;
- worker de expiração e devolução de estoque;
- confirmação de pagamento;
- PostgreSQL no caminho síncrono;
- outbox ou mensageria;
- Redis Cluster real;
- emissão de ingressos.

Esses limites são parte da aprendizagem: a Aula 3 torna o problema concreto; a Aula 4 transforma os passos críticos numa operação atômica.

## Referências para consulta durante a aula

- Contrato HTTP do projeto: `contracts/openapi.yaml`
- Plano de estudos: `docs/plano-de-estudos.md`
- Redis `SET`: https://redis.io/docs/latest/commands/set/
- Redis hashes: https://redis.io/docs/latest/develop/data-types/hashes/
- Redis sorted sets: https://redis.io/docs/latest/develop/data-types/sorted-sets/
- Contexto em Go: https://pkg.go.dev/context
