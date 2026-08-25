# Plano de estudos: Go + Redis

## Como trabalharemos

Cada módulo segue o mesmo ciclo:

1. estudar um conceito pequeno;
2. implementar uma mudança curta;
3. escrever ou completar testes;
4. executar um experimento;
5. explicar com suas palavras por que a solução funciona;
6. receber revisão antes de avançar.

Não copie uma implementação completa. Quando estiver bloqueado, traga o código, o erro e sua hipótese. A revisão deve ensinar o raciocínio, não apenas entregar a resposta.

## Trilha

### Módulo 0 — Ambiente e primeiro servidor

**Conceitos:** módulo Go, pacotes, `net/http`, contexto, encerramento gracioso e testes HTTP.

**Entrega:** `GET /health` conforme o contrato OpenAPI, inicialmente sem Redis.

**Critério:** `go test ./...` passa e `curl http://localhost:8080/health` retorna HTTP 200.

### Módulo 1 — Domínio antes da infraestrutura

**Conceitos:** entidades, invariantes, erros de domínio e testes de tabela.

**Entrega:** entidade `Event` com capacidade e preço válidos; criação e listagem ainda em memória.

**Critério:** entradas inválidas não criam eventos e erros HTTP seguem o contrato.

### Módulo 2 — Redis como estado operacional

**Conceitos:** tipos de dados Redis, serialização, nomes de chaves, TTL, conexão e timeouts.

**Entrega:** catálogo de eventos e disponibilidade armazenados no Redis.

**Critério:** reiniciar somente a API não perde os eventos e uma indisponibilidade do Redis resulta em HTTP 503 controlado.

### Módulo 3 — Reservas e idempotência

**Conceitos:** concorrência, read-modify-write, chave de idempotência e consistência.

**Entrega:** criação de reserva com duração de dez minutos.

**Critério:** repetir a mesma requisição com a mesma `Idempotency-Key` devolve a mesma reserva sem consumir estoque novamente.

### Módulo 4 — Atomicidade no Redis

**Conceitos:** scripts Lua/Redis Functions, execução atômica e modelagem de chaves por agregado.

**Entrega:** verificar estoque, reservar, registrar idempotência e agendar expiração em uma única operação atômica.

**Critério:** o teste de carga nunca deixa o estoque negativo nem aceita mais ingressos que a capacidade.

### Módulo 5 — Expiração confiável

**Conceitos:** sorted sets, workers, leases, repetição segura e relógios.

**Entrega:** worker que encontra reservas vencidas e devolve seus ingressos atomicamente.

**Critério:** interromper e reiniciar o worker não perde expirações nem devolve estoque duas vezes.

### Módulo 6 — Assincronismo e outbox

**Conceitos:** dual write, outbox, entrega pelo menos uma vez e consumidores idempotentes.

**Entrega:** registrar eventos de domínio numa outbox Redis no mesmo comando da reserva.

**Critério:** falhas simuladas não criam uma reserva sem o respectivo evento de outbox.

### Módulo 7 — Particionamento space-based

**Conceitos:** afinidade por `event_id`, hash tags, hot spots e operações entre shards.

**Entrega:** convenção de chaves `{event_id}:...` preparada para Redis Cluster.

**Critério:** todas as chaves usadas por uma operação atômica pertencem ao mesmo hash slot.

### Módulo 8 — Observabilidade e carga

**Conceitos:** logs estruturados, métricas de negócio, percentis, saturação e backpressure.

**Entrega:** métricas de reservas aceitas/recusadas, latência e disponibilidade por evento.

**Critério:** explicar, com dados de um teste k6, o primeiro gargalo encontrado.

## Regras arquiteturais

- PostgreSQL não participa do caminho síncrono de reserva.
- Redis não é somente cache; representa o estado operacional do espaço.
- `event_id` é a chave de afinidade.
- Toda mutação externa deve ser idempotente.
- Falhas de infraestrutura são traduzidas para erros HTTP explícitos.
- Concorrência é verificada por teste, não presumida.

