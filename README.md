# Sistema de gestão de ingressos

Laboratório de estudo de arquitetura space-based. O projeto separa deliberadamente as responsabilidades:

- **Aluno:** API Go, domínio, integração e operações atômicas no Redis.
- **Fundação pronta:** interface React, PostgreSQL, Docker Compose, contrato HTTP e exercícios verificáveis.

O backend ainda não está implementado. Isso é intencional: cada parte dele será construída como uma aula prática.

## Subindo o ambiente

Pré-requisitos atuais: Docker e Docker Compose. Para desenvolver o backend também será necessário instalar Go 1.27 ou uma versão compatível.

```bash
cp .env.example .env
docker compose up --build -d
docker compose ps
```

Serviços:

- Frontend: http://localhost:3100
- Redis: `localhost:6379`
- PostgreSQL: `localhost:5433`
- Futuro backend Go: `http://localhost:8080`

Por padrão o frontend usa dados simulados (`VITE_API_MODE=mock`). Quando o primeiro endpoint Go estiver pronto, altere para:

```dotenv
VITE_API_MODE=http
```

Depois reconstrua o frontend:

```bash
docker compose up --build -d frontend
```

## Material de estudo

- [Plano de estudos](docs/plano-de-estudos.md)
- [Aula 0 — primeiro servidor Go](.aulas/AULA-00.md)
- [Aula 1 — domínio de eventos](.aulas/AULA-01.md)
- [Aula 2 — Redis como estado operacional](.aulas/AULA-02.md)
- [Aula 3 — reservas e idempotência](.aulas/AULA-03.md)
- [Contrato OpenAPI](contracts/openapi.yaml)
- [Schema de persistência](database/init/001_schema.sql)
- [Teste de concorrência](tests/load/reservas.js)

As anotações pessoais permanecem em [anotacoes-igor.md](anotacoes-igor.md).
