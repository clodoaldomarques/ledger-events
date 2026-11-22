# Plano de Implementação: API de Health Checks

**Branch**: `002-api-de-health` | **Data**: 2026-11-22 | **Spec**: [spec.md](spec.md)

**Entrada**: especificação derivada de `internal/adapters/inbound/http/server`, `internal/adapters/outbound/mysql`, `internal/adapters/outbound/sns` e `docs/specs/openapi.yaml`.

**Nota**: este plano descreve decisões **já presentes** no código; não é um plano a executar do zero.

## Resumo

Registrar duas rotas no Echo e responder com `c.NoContent`. A liveness é independente; a readiness compõe `Repository.Health()` e `Topic.Health()` com adapters criados por chamada e um contexto de 500ms.

## Contexto Técnico

**Linguagem/Versão**: Go 1.26.1 (`go.mod`)

**Dependências principais**: Echo v4.16.0, core-sdk v0.8.8 (publisher SNS), `go-sql-driver/mysql`

**Armazenamento**: MySQL `ledger`, via `config.New().GetMySQLConnectionString()`

**Testes**: `go test ./...`; sondas verificadas manualmente com `curl` — não há teste automatizado dedicado

**Plataforma alvo**: serviço HTTP em container, porta `5001`

**Tipo de projeto**: serviço HTTP (rotas de infraestrutura)

**Metas de desempenho**: prontidão limitada a `500ms` por chamada

**Restrições**: sem corpo de resposta, sem autenticação, adapters recriados por chamada

**Escopo**: 2 rotas HTTP, 2 verificações de dependência

## Verificação da Constituição

*GATE: aplicado antes de Phase 0 e reavaliado após o desenho.*

| Princípio | Situação | Evidência |
|---|---|---|
| I. Separação em camadas | Passa | Rotas e handlers em `internal/adapters/inbound/http/server` |
| II. O domínio calcula, a interface transporta | N/A | Sondas não contêm regras de domínio |
| III. Configuração singleton | Passa | `config.New()` usado pelos construtores dos adapters |
| IV. Observabilidade no request | Passa | Interceptadores registrados em `server.routes()` |
| V. Persistência transacional antes da publicação | N/A | Sondas não gravam nem publicam eventos |
| VI. Testes unitários determinísticos | Atenção | Sem teste automatizado das rotas de health |
| VII. Infraestrutura como código | Passa | Dependências declaradas em `docker-compose.yaml` |

## Estrutura do Projeto

### Documentação (esta funcionalidade)

```text
docs/specs/002-api-de-health/
├── spec.md
├── plan.md
└── tasks.md
```

### Código-fonte (raiz do repositório)

```text
internal/adapters/inbound/http/server/server.go        # routes(), Liveness, Readiness
internal/adapters/outbound/mysql/repository.go       # Health(), Close()
internal/adapters/outbound/sns/service.go           # Health(), Close(ctx)
config/config.go                            # GetMySQLConnectionString()
docs/specs/openapi.yaml                     # /health/live e /health/ready
```

**Decisão de estrutura**: as sondas vivem no servidor, não no handler de domínio; a prontidão instancia seus próprios adapters em vez de reaproveitar os do caso de uso.

## Detalhamento do fluxo

```text
GET /health/live
  └── Server.Liveness → c.NoContent(200)

GET /health/ready
  └── Server.Readiness
        ├── context.WithTimeout(request.Context(), 500ms)
        ├── mysql.NewRepository(ctx) → Health() (PingContext) → 503 se falhar
        ├── message.New(ctx).Health() (GetTopicAttributes) → 503 se falhar
        └── c.NoContent(200)
```

`defer db.Close()` é chamado para o repositório. O publicador SNS é criado em linha (`message.New(ctx).Health()`) e não recebe `Close(ctx)` explícito nessa rota.

## Complexidade Conhecida

| Desvio | Por que existe | Alternativa mais simples considerada |
|---|---|---|
| Adapters por chamada de prontidão | Mesmo padrão de composição do handler de eventos | Injetar `HealthChecker` compartilhado (avaliado na spec 005) |
| `db == nil` em falha de conexão | `NewRepository` não propaga erro (`OBS-02`) | Retornar erro do construtor antes do ping |
| Sem teste automatizado das sondas | Verificação manual com `curl` | Usar `httptest` com fakes de `Health()` |
| Timeout fixo de 500ms | Valor literal no handler | Movê-lo para configuração (fora do escopo desta spec) |
