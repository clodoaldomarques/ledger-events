# Plano de Implementação: Adoção da Arquitetura Hexagonal — `ledger-events`

**Branch**: `005-hexagonal-adapters` | **Data**: 2026-11-22 | **Repo**: `github.com/clodoaldomarques/ledger-events`

**Entrada**: [spec.md](spec.md), `docs/specs/constitution.md`, `docs/specs/platform.md`, specs `001`-`004` e a estrutura atual deste repo.

**Nota**: este plano descreve trabalho **a executar** neste repo. `[x]` só deve ser usado após a implementação real, nunca por ter escrito a spec. Todo o conteúdo necessário está aqui e nos arquivos vizinhos deste diretório.

## Resumo

Reorganizar o módulo Go de `ledger-events` para uma topologia hexagonal explícita: `internal/core/ledger` (domínio + aplicação + portas), `internal/adapters` (entrada HTTP e saída MySQL/Ledger Config/SNS) e `cmd/main.go` como composition root. O handler deixa de construir três adapters por requisição e `Readiness` passa a usar uma porta injetada. A dependência legada `github.com/golang/mock` e os erros de domínio não usados são removidos. Regras, contrato HTTP, envelope SNS, DDL, spans e desvios documentados ficam congelados.

## Contexto Técnico

**Linguagem/Versão**: Go 1.26.1 (`go.mod`); Dockerfile usa Go + Alpine

**Dependências principais**: Echo v4.16.0, core-sdk v0.8.8 (tracer/logger/métricas/HTTP/SNS), `go-sql-driver/mysql`, `sony/gobreaker`, `shopspring/decimal`, `expr-lang/expr`, `google/uuid`, `testify`, `go.uber.org/mock` (mockgen) e `github.com/golang/mock` (legado, a remover)

**Armazenamento**: MySQL `ledger`, tabelas `event` e `entry`, `decimal(20,15)` — nenhuma alteração de schema

**Testes unitários**: `go test ./...` (somente mocks/gomock e `httptest`, sem Docker/MySQL/LocalStack). `make test` escreve `cover.out` e abre o HTML — não usar em checagens não interativas.

**Benchmarks**: `internal/application/events/bench_test.go` — `BenchmarkCreateEvent_Success`, `BenchmarkCreateEvent_Parallel`; devem continuar compilando após o movimento.

**Mocks**: `go generate ./internal/application/events`; `mock.go` é gerado a partir de `interfaces.go` — editar a interface e regenerar, nunca editar o arquivo gerado.

**Plataforma alvo**: serviço HTTP em container (porta default 5001 via `APP_PORT`)

**Restrições**: sem CI, sem autenticação, `config.New()` é singleton `sync.Once`, tracing dentro do core permanece nesta fase, `x-cid` é auto-gerado quando ausente, timeout de prontidão 500ms e de shutdown 25s permanecem.

**Escopo neste repo**: 1 movimento de pacotes, 4 portas de saída + 1 de entrada, 2 adaptadores de entrada (handler + servidor) convertidos, 3 adaptadores de saída realocados, 1 guarda de arquitetura, 1 remoção de código morto.

## Verificação da Constituição

*GATE: aplicado antes da Fase 0 e reavaliado ao fim da Fase 6.*

| Princípio | Situação planejada | Evidência alvo neste repo |
|---|---|---|
| I. Separação em camadas | **Reescrito** para "regra de dependência via portas" | `internal/core/ledger/**` sem imports de adapters/AWS/Echo/config; `cmd/main.go` como composition root |
| II. O domínio calcula, a interface transporta | Passa | DTOs permanecem em `adapters/inbound/http`; cálculo em `ProcessRegular`/`ProcessMigration` |
| III. Configuração por ambiente, com singleton explícito | Passa (inalterado) | `config.New()` continua `sync.Once`; adapters recebem config na construção |
| IV. Observabilidade é parte do request | Passa (com desvio documentado) | Spans permanecem em casos de uso/adapters; `x-cid`/`x-tenant` continuam no adaptador HTTP |
| V. Persistência transacional antes da publicação | Passa | Ordem config → cálculo → validação → persistência → publicação inalterada |
| VI. Testes unitários determinísticos | Passa | Mocks gerados ao lado das portas; fakes in-memory; suíte sem serviços externos |
| VII. Infraestrutura como código | Passa | Compose/Terraform/k8s/Dockerfile sem mudança; `make arch-check` adicionado |

**Ação obrigatória na Fase 6**: reescrever o Princípio I de `docs/specs/constitution.md` e atualizar sua data/versão.

## Estrutura do Projeto

### Documentação (esta funcionalidade, neste repo)

```text
docs/specs/005-hexagonal-adapters/
├── spec.md      # US1-US5, FR-001..FR-014, SC-001..SC-007 aplicados a ledger-events
├── plan.md      # este documento: layout, mapeamento, riscos, verificação
└── tasks.md     # fases 0-6 com IDs T001-T031 e LE-001-LE-028
```

### Código-fonte alvo

```text
cmd/main.go                              # composition root
config/config.go                         # mantido; sync.Once inalterado
internal/
├── core/ledger/
│   ├── domain/                          # event.go, entry.go, configs.go, processing.go,
│   │                                    # regular.go, migration.go, errors.go
│   ├── application/                     # services.go, services_test.go, bench_test.go
│   └── ports/
│       ├── inbound.go                   # EventUseCase
│       ├── outbound.go                  # EventRepository, EventConfigProvider,
│       │                                # EventPublisher, HealthChecker (+ go:generate)
│       └── mock.go                      # gerado
└── adapters/
    ├── inbound/http/
    │   ├── ledger/                      # handler.go, request.go, response.go
    │   ├── server/                      # Echo, rotas, /health/ready
    │   └── shared/                      # x-cid/x-tenant, erros.go, EntityRequest
    └── outbound/
        ├── mysql/                       # EventRepository + HealthChecker
        ├── config/                      # EventConfigProvider (HTTP + circuit breaker)
        └── sns/                         # EventPublisher + HealthChecker
scripts/check-architecture.sh            # guarda de imports
```

**Decisão de estrutura**: topologia horizontal (`core`/`adapters`) em vez de vertical por contexto, para tornar a regra de dependência visível no caminho do arquivo e manter o mesmo padrão dos demais repos da família.

## Mapeamento de Pastas

| Origem | Destino | Observações |
|---|---|---|
| `internal/domain/events` | `internal/core/ledger/domain` | Somente `git mv` + imports; nenhuma regra muda |
| `internal/application/events/services.go` | `internal/core/ledger/application/services.go` | `services_test.go` e `bench_test.go` junto |
| `internal/application/events/interfaces.go` | `internal/core/ledger/ports/outbound.go` | Diretiva `go:generate` vai junto; criar `inbound.go` |
| `internal/application/events/mock.go` | `internal/core/ledger/ports/mock.go` | Regenerar, nunca editar à mão |
| `internal/infra/rest/ledger` | `internal/adapters/inbound/http/ledger` | Handler vira struct com `EventUseCase` na Fase 3 |
| `internal/infra/rest/server` | `internal/adapters/inbound/http/server` | `server.New()` passa a receber handler e `HealthChecker` |
| `internal/infra/rest/shared` | `internal/adapters/inbound/http/shared` | Mantém `x-cid`/`x-tenant`, `EntityRequest` e mapeador de erros |
| `internal/infra/db/mysql` | `internal/adapters/outbound/mysql` | `connection.go`/`config` sobem para a raiz; cliente injetado |
| `internal/infra/ledger/config` | `internal/adapters/outbound/config` | Cliente HTTP e circuit breaker configurados na raiz |
| `internal/infra/message` | `internal/adapters/outbound/sns` | `message.New(ctx)` vira `sns.New(publisher)` |

## Riscos e Decisões

| Risco / desvio | Por que existe | Decisão |
|---|---|---|
| `tracer`/`logger` dentro do core | Já presentes; mover para decorators exigiria reescrever spans e testes | Manter e registrar como desvio; decorators em spec separada |
| `config.New()` singleton | Comportamento atual documentado | Manter; apenas parar de chamá-lo dentro dos adapters |
| `sns.Publisher.Health()` usa o `ctx` recebido no construtor | Compartilhar o publisher criaria health check sem o prazo de 500ms do `/health/ready` | O adapter SNS guarda a config e cria um publisher somente para `Health(ctx)`, preservando o prazo; o publisher de `Emit` continua único |
| Construtores legados (`NewFromConfig`) na Fase 2 | Tornar a Fase 2 verificável sem alterar a composição da Fase 3 | Wrappers transitórios deletados na Fase 3, quando `cmd/main.go` passa a injetar os clientes |
| `HealthChecker` no adapter de config | `/health/ready` hoje consulta apenas MySQL e SNS | Não implementar GET de config: isso mudaria o contrato de prontidão (spec própria) |
| De "por request" para "por processo" | Exige composição única na raiz | Confirmado na Fase 0: `sql.DB` e `aws-sdk-go-v2` são seguros para uso concorrente; `httpclient.Client()` do core-sdk usa `http.DefaultTransport` como base. Instâncias compartilhadas na raiz, sem factory |
| Mover diretórios quebra todos os imports de uma vez | Mudança física | Fase 1 exclusivamente `git mv` + imports, com build verde antes de seguir |
| Portas genéricas (`Repository`, `ConfigProvider`, `Topic`) | Nomes históricos | Renomear na Fase 2; aceitar quebra de imports internos |
| Interfaces `Config`/`Script` implementadas por DTOs de infra | DIP atual, mas DTOs vivem em `infra/ledger/config` | Mover os DTOs com o adapter; manter as interfaces no domínio |
| Bugs catalogados nas specs 001-004 | Registrados como comportamento atual | Congelar; cada correção com sua própria spec |
| Collection Postman desatualizada | Contrato antigo (`OBS-08`) | Usar `openapi.yaml` como referência nesta spec |
| `make test` abre navegador | `go tool cover -html` | Usar `go test ./...` / `-coverprofile` em verificações |

## Estratégia de Migração

1. **Fase 0 — Baseline e guarda**: suíte verde e `make arch-check` passando já no estado atual.
2. **Fase 1 — Movimento físico**: `git mv` + ajuste de imports e `go:generate`; regenerar mocks. Nenhuma mudança de comportamento.
3. **Fase 2 — Portas**: renomear e posicionar portas; criar `EventUseCase`; construtores recebem cliente/config.
4. **Fase 3 — Composition root**: `cmd/main.go` compõe tudo; handler vira struct; `/health/ready` e erros passam por portas/mapper.
5. **Fase 4 — Limpeza**: migrar teste para `go.uber.org/mock`, remover erros não usados e `github.com/golang/mock`; `go mod tidy`.
6. **Fase 5 — Opcional (fora do SC)**: decorators de tracing, divisão de portas por capacidade, correções de bugs conhecidos.
7. **Fase 6 — Fechamento**: constituição, `platform.md`, índice de specs e planos `001`-`004`; verificações finais.

Cada fase DEVE terminar com build e testes verdes. Rollback é feito por commit de fase; nunca misturar fases diferentes no mesmo commit.

## Verificação

```bash
# executar na raiz do repo

go test ./...
go build ./...
go vet ./...
make arch-check

# após o ajuste de imports nos testes
go test ./internal/core/ledger/application -run TestService_CreateEvent -count=1
go generate ./internal/core/ledger/...
go mod tidy

# rastreabilidade do código morto
grep -R -E 'golang/mock|ErrDuplicatedEntry|ErrDuplicatedAccount' --include='*.go' --include='go.mod' . || true

# contrato externo (sem dependências externas)
# comparar respostas de POST /v1/ledger/events e /health/* com docs/specs/openapi.yaml
```

## Dependências e Ordem de Execução

- A spec deve ser aprovada antes da Fase 0.
- A spec canônica da família está em `ledger-config/docs/specs/005-hexagonal-adapters/`; este repo replica os IDs e a sequência de fases.
- Fase 0 bloqueia todas as demais; Fase 1 deve estar verde antes da Fase 2.
- Fase 2 bloqueia Fase 3; Fase 4 pode ser paralela à Fase 3.
- Fase 5 depende de decisão explícita e não é necessária para fechar a spec.
- A Fase 6 pode ocorrer antes da Fase 3 de outro repo, desde que a constituição aqui reflita o estado real.
