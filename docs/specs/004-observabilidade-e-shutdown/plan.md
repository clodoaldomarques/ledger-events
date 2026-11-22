# Plano de Implementação: Observabilidade e Desligamento Gracioso

**Branch**: `004-observabilidade-e-shutdown` | **Data**: 2026-11-22 | **Spec**: [spec.md](spec.md)

**Entrada**: especificação derivada de `cmd/main.go`, `internal/infra/rest/server`, `internal/application/events`, `internal/infra/` e `scripts/docker/otel-collector.yaml`.

**Nota**: este plano descreve decisões **já presentes** no código; não é um plano a executar do zero.

## Resumo

Telemetria inicializada no startup via `otel.Start(ctx)`, interceptadores registrados no Echo e spans criados manualmente no caso de uso e nos adapters com o atributo `cid`. O `main` aguarda `SIGINT`/`SIGTERM`, usa contexto de `25s`, encerra o HTTP e depois o OTel. Erros são impressos no stdout e não alteram o status de saída.

## Contexto Técnico

**Linguagem/Versão**: Go 1.26.1 (`go.mod`)

**Dependências principais**: core-sdk v0.8.8 (`pkg/otel`, `pkg/zap/logger`, `pkg/otel/echo`, `pkg/otel/tracer`), OpenTelemetry 1.47, Echo v4.16.0

**Armazenamento**: `event.correlation_id` é a única persistência do `cid`

**Testes**: `go test ./...` sem coletor; telemetria verificada manualmente no ambiente local

**Plataforma alvo**: serviço Linux em container, OTLP gRPC em `localhost:4317`

**Tipo de projeto**: serviço HTTP com telemetria cross-cutting

**Metas de desempenho**: não declaradas

**Restrições**: timeout de shutdown fixo; `logger.Fatal` em erro de `Start` diferente de `http.ErrServerClosed`; tracing dentro do core permanece nesta fase

**Escopo**: 3 interceptadores, 4 spans, 1 ciclo de startup/shutdown

## Verificação da Constituição

| Princípio | Situação | Evidência |
|---|---|---|
| I. Separação em camadas | Passa | `main`, servidor e adapters em pacotes distintos |
| II. O domínio calcula, a interface transporta | Passa | Spans são instrumentação, não regra de negócio |
| III. Configuração singleton | Passa | `config.New()` usado para porta e variáveis OTEL |
| IV. Observabilidade no request | Passa | Interceptadores em `server.routes()` e spans com `cid` |
| V. Persistência transacional antes da publicação | N/A | Fora do escopo desta spec |
| VI. Testes unitários determinísticos | Atenção | Telemetria não é coberta por teste automatizado |
| VII. Infraestrutura como código | Passa | Coletor, Zipkin e Prometheus em `docker-compose.yaml` |

## Estrutura do Projeto

### Documentação (esta funcionalidade)

```text
docs/specs/004-observabilidade-e-shutdown/
├── spec.md
├── plan.md
└── tasks.md
```

### Código-fonte (raiz do repositório)

```text
cmd/main.go                                  # otel.Start, sinais, Shutdown, otel.Shutdown
internal/infra/rest/server/server.go         # interceptadores e rotas
internal/infra/rest/ledger/handler.go        # x-cid/x-tenant → cid do serviço
internal/application/events/services.go      # span Service::CreateEvent
internal/infra/ledger/config/ledger_config.go# span LedgerConfigApi::FindConfigByLevel
internal/infra/db/mysql/repository.go        # span Repository::SaveEvent
internal/infra/message/service.go            # span Topic::Emit + Close(ctx)
scripts/docker/otel-collector.yaml
```

## Detalhamento do fluxo

```text
main
 ├── otel.Start(ctx)
 ├── server.New() → routes(): tracer → logger → otecho → handlers
 ├── go server.Start(APP_PORT)
 ├── signal.Notify(SIGINT, SIGTERM) → <-quit
 ├── ctx, cancel := context.WithTimeout(ctx, 25s)
 ├── Server.Shutdown(ctx)     → imprime erro/sucesso no stdout
 ├── otel.Shutdown(ctx)       → imprime erro no stdout, se houver
 └── fmt.Println("Graceful shutdown concluído") → exit 0
```

Spans e atributos:

| Span | Local | Atributo de correlação |
|---|---|---|
| `Service::CreateEvent` | `application/events/services.go` | `cid` |
| `LedgerConfigApi::FindConfigByLevel` | `infra/ledger/config/ledger_config.go` | `cid` |
| `Repository::SaveEvent` | `infra/db/mysql/repository.go` | `cid` |
| `Topic::Emit` | `infra/message/service.go` | `cid` |

## Complexidade Conhecida

| Desvio | Por que existe | Alternativa mais simples considerada |
|---|---|---|
| Tracing chamado dentro do core | Padrão histórico do core-sdk | Mover para decorators nas bordas (spec 005, fase opcional) |
| Timeout de 25s literal no `main` | Sem configuração dedicada | Expô-lo como variável de ambiente |
| Erros de shutdown apenas impressos | `main` não retorna erro | Propagar status de saída não zero |
| `Topic.Close(ctx)` só loga | Publicador do core-sdk não exige fechamento | Remover o método ou fechar o recurso real |
| Sem teste de telemetria | Dependência de coletor | Usar tracer noop em `TestMain`, como em outros repos da família |
