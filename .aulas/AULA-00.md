# Aula 0 — seu primeiro servidor Go

Nesta aula você implementará somente o endpoint de saúde. Não conecte ao Redis ainda.

## 1. Preparar o ambiente

O comando abaixo precisa funcionar:

```bash
go version
```

No momento da criação do projeto, a versão estável mais recente é Go 1.27. Depois de instalar, entre neste diretório e inicialize um módulo usando um caminho que represente o seu repositório:

```bash
cd backend
go mod init github.com/SEU_USUARIO/sistema-gestao-ingressos/backend
```

Não use literalmente `SEU_USUARIO`: substitua pelo seu usuário ou por outro caminho de módulo que você controle.

## 2. Sua tarefa

Crie a menor estrutura possível:

```text
backend/
├── cmd/api/main.go
└── internal/httpapi/health.go
```

Implemente:

```http
GET /health
```

Resposta esperada nesta primeira aula:

```json
{
  "status": "ok",
  "redis": "not_checked",
  "version": "dev"
}
```

Requisitos:

- usar apenas a biblioteca padrão;
- servidor na porta `8080`;
- header `Content-Type: application/json`;
- método diferente de `GET` deve retornar `405`;
- adicionar um teste com `httptest`;
- configurar timeouts do `http.Server`;
- tratar `SIGINT` e `SIGTERM` com encerramento gracioso.

## 3. Como verificar

```bash
go fmt ./...
go test ./...
go vet ./...
go run ./cmd/api
```

Em outro terminal:

```bash
curl -i http://localhost:8080/health
```

Quando terminar, peça uma revisão antes de começar o Redis. Na revisão, esteja preparado para explicar:

1. por que o handler recebe `http.ResponseWriter` e `*http.Request`;
2. por que o JSON deve ser serializado com `encoding/json`;
3. para que servem os timeouts do servidor;
4. o que acontece com requisições em andamento durante o shutdown.

