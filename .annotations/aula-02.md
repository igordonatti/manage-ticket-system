**Redis ttl command return -2:** In Redis, a return value of -2 from the TTL or PTTL command means that the specified key does not exist.
**Redis ttl command return -1:** The key exists, but is has no associated expiration time (it persists indefinitaly)

---

1. **Por que construir `redis.Client` não garante uma conexão funcional?**
Construir o cliente apenas configura o acesso ao Redis. A conexão funcional só é comprovada quando executamos um comando, como PING.

2. **Por que usar `r.Context()?`**
O `context.Context` representa o ciclo de vida de uma operação.
No servidor HTTP, o Go cria um contexto para cada requisição:
`r.Context()`
Ele pode ser cancelado quando:
- o cliente fecha a conexão;
- a requisição é cancelada;
- o handler termina;
- algum prazo superior expira.

O contexto do PING nasce de `r.Context()` para herdar cancelamentos e prazos da requisição HTTP, evitando continuar trabalhando depois que já foi cancelada.

Se usássemos `context.Background()` o PING ficaria desconectado da requisição, a operação poderia continuar desnecessariamente.

---

Comando                Responsabilidade
━━━━━━━━━━━━━━━━━━━━━  ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
go fmt ./...           Padronizar a formatação
─────────────────────  ───────────────────────────────────────────────────────────
go test ./...          Compilar e executar testes
─────────────────────  ───────────────────────────────────────────────────────────
go test -race ./...    Executar testes procurando acessos concorrentes inseguros
─────────────────────  ───────────────────────────────────────────────────────────
go vet ./...           Analisar construções suspeitas no código

---

