# Aula 1 — domínio de eventos

Nesta aula o evento nasce sem conhecer HTTP, JSON, Redis ou PostgreSQL. O objetivo é aprender a colocar regras no domínio antes de escolher como os dados entram ou são armazenados.

## Pré-checkpoint da Aula 0

Antes de iniciar esta aula, faça duas correções na implementação anterior:

1. Inverta o controle do shutdown: execute `ListenAndServe` numa goroutine, espere o sinal em `main` e chame `Shutdown` a partir de `main`. O processo principal deve aguardar `Shutdown` terminar.
2. No teste do health handler, decodifique o corpo JSON e verifique os três campos esperados. Um teste que verifica somente status e header aceitaria um corpo incorreto.

Depois execute novamente:

```bash
go fmt ./...
go test -race ./...
go vet ./...
```

## Conceitos desta aula

- struct e métodos;
- campos exportados e não exportados;
- construtor idiomático em Go;
- invariantes;
- erros de domínio;
- testes orientados a comportamento;
- separação entre entidade e representação HTTP.

Go não possui construtores como parte da linguagem. Chamamos de construtor uma função como `New(...)` que cria um valor somente depois de verificar suas regras.

## Checkpoint 1A — entidade Event

Crie somente:

```text
internal/event/
├── event.go
└── event_test.go
```

A entidade deve possuir conceitualmente:

- identificador textual;
- nome;
- local;
- instante do evento;
- início das vendas;
- capacidade;
- preço em centavos;
- status.

### Regras obrigatórias

- identificador, nome e local não podem ficar vazios depois de remover espaços das pontas;
- nome e local devem ter ao menos três caracteres;
- capacidade deve ser maior que zero;
- preço em centavos pode ser zero, mas nunca negativo;
- início das vendas deve ocorrer antes do evento;
- um evento novo começa com status `draft`;
- nome e local devem ser armazenados sem espaços extras nas pontas.

Não compare as datas com `time.Now()`: isso tornaria o domínio dependente do relógio da máquina e os testes menos determinísticos.

### Decisões esperadas

- Use um tipo próprio para o status, em vez de strings espalhadas.
- Use `int64` para preço em centavos; não use `float64` para dinheiro.
- Prefira campos não exportados para impedir que outra parte do programa viole as regras depois da criação.
- Forneça métodos de leitura com nomes como `Name()` e `Capacity()`, sem prefixo `Get`.
- Retorne erros explícitos para entradas inválidas. Você pode usar erros sentinela com `errors.New`.
- Não adicione tags JSON à entidade. A futura camada HTTP terá seus próprios DTOs.

## Testes mínimos

Escreva testes de tabela cobrindo:

1. evento válido;
2. preço gratuito, igual a zero;
3. identificador vazio;
4. nome vazio ou curto;
5. local vazio ou curto;
6. capacidade zero e negativa;
7. preço negativo;
8. vendas começando no mesmo instante do evento;
9. vendas começando depois do evento;
10. remoção dos espaços externos de nome e local;
11. status inicial igual a `draft`.

Use datas fixas nos testes, por exemplo:

```go
salesStart := time.Date(2026, time.September, 1, 10, 0, 0, 0, time.UTC)
eventStart := time.Date(2026, time.October, 17, 23, 0, 0, 0, time.UTC)
```

## Limites do exercício

Neste checkpoint, não implemente:

- handler HTTP;
- armazenamento em memória;
- Redis;
- geração de UUID;
- atualização de evento;
- abertura de vendas.

Esses limites mantêm o problema pequeno o bastante para avaliarmos se as invariantes estão realmente protegidas.

## Perguntas para a revisão

Quando terminar, responda:

1. Qual é a diferença prática entre validar no handler e validar no construtor da entidade?
2. Por que `float64` é uma escolha ruim para representar preço?
3. O que campos não exportados protegem e o que eles não protegem?
4. Por que os testes usam datas fixas em vez de `time.Now()`?

O checkpoint 1B adicionará um repositório em memória e os endpoints de criação e listagem somente depois da revisão desta entidade.

