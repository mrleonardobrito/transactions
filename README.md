# Transactions Service

Módulo de **transações bancárias** de uma arquitetura em microsserviços voltada a uma aplicação de **monitoramento financeiro**.

Este serviço é responsável por registrar e listar as movimentações financeiras (receitas e despesas) dos usuários. Ele expõe uma API HTTP e persiste os dados em um banco PostgreSQL compartilhado pelo domínio financeiro.

## Arquitetura

Este repositório representa **um único microsserviço** dentro de um ecossistema maior de monitoramento financeiro. Sua responsabilidade é limitada ao domínio de **transações** (`operacoes`), enquanto outros serviços cuidam de domínios como autenticação, usuários, relatórios, etc.

- **Linguagem:** Go 1.23
- **Banco de dados:** PostgreSQL 16
- **Documentação:** Swagger (via `swaggo`)
- **Porta:** `8080`

## Rotas

A API expõe o recurso `/transactions`. A documentação interativa (Swagger UI) fica disponível em `http://localhost:8080/docs/index.html`.

### `GET /transactions`

Lista todas as transações, ordenadas pela data de criação (mais recentes primeiro).

**Resposta `200 OK`**

```json
[
  {
    "id": 1,
    "usuario_id": 2,
    "tipo": "receita",
    "valor": 1500.50,
    "descricao": "Salário",
    "data_criacao": "2026-06-23T16:00:00Z"
  }
]
```

| Código | Descrição |
| ------ | --------- |
| `200`  | Lista de transações retornada com sucesso |
| `500`  | Erro interno ao buscar transações |

### `POST /transactions`

Registra uma nova transação financeira.

**Corpo da requisição**

```json
{
  "usuario_id": 2,
  "tipo": "receita",
  "valor": 1500.50,
  "descricao": "Salário"
}
```

> O campo `tipo` deve ser obrigatoriamente `"receita"` ou `"despesa"`. O campo `descricao` é opcional.

**Resposta `201 Created`**

```json
{
  "id": 1,
  "usuario_id": 2,
  "tipo": "receita",
  "valor": 1500.50,
  "descricao": "Salário",
  "data_criacao": "2026-06-23T16:00:00Z"
}
```

| Código | Descrição |
| ------ | --------- |
| `201`  | Transação criada com sucesso |
| `400`  | Dados inválidos (JSON malformado ou `tipo` diferente de `receita`/`despesa`) |
| `500`  | Erro interno ao salvar a transação |

## Como instalar e rodar

### Pré-requisitos

- [Go 1.23+](https://go.dev/dl/)
- [Docker](https://www.docker.com/) e Docker Compose

### 1. Suba o banco de dados

O banco PostgreSQL é provisionado via Docker Compose, com as tabelas e os dados de exemplo carregados automaticamente na primeira inicialização.

```bash
cd financeiro_db
docker compose up -d
```

O banco ficará disponível em `localhost:5434` com as credenciais:

| Variável | Valor |
| -------- | ----- |
| Usuário  | `financeiro` |
| Senha    | `financeiro` |
| Database | `financeiro` |

### 2. Rode o serviço

A string de conexão pode ser definida pela variável de ambiente `DATABASE_URL`. Caso não seja informada, o serviço usa por padrão a configuração do Docker Compose acima.

```bash
# Da raiz do projeto
go mod download
go run .
```

Ou, definindo explicitamente a conexão:

```bash
DATABASE_URL="postgres://financeiro:financeiro@localhost:5434/financeiro?sslmode=disable" go run .
```

O servidor inicia em `http://localhost:8080`.

### Rodando com Docker

A aplicação também pode ser empacotada e executada via Docker:

```bash
docker build -t transactions-service .
docker run -p 8080:8080 \
  -e DATABASE_URL="postgres://financeiro:financeiro@host.docker.internal:5434/financeiro?sslmode=disable" \
  transactions-service
```

## Estrutura do projeto

```
.
├── main.go               # Servidor HTTP e handlers das rotas
├── docs/                 # Documentação Swagger gerada
├── Dockerfile
└── go.mod
```
