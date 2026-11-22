# 01 — Requisitos

Todos os requisitos abaixo são `AS-IS`, salvo indicação contrária.

## Requisitos funcionais

| ID | Requisito | Critério de aceitação | Rastreabilidade |
|---|---|---|---|
| FR-01 | O serviço deve aceitar `POST /v1/ledger/events` com corpo JSON. | Rota registrada no servidor Echo; corpo vinculado a `EventRequest`. | `internal/infra/rest/server/server.go`, `internal/infra/rest/ledger/handler.go` |
| FR-02 | Toda requisição deve possuir um correlation id. | Usa `x-cid` quando presente; caso ausente, gera `uuid.NewString()`. | `handler.go::getHeaders` |
| FR-03 | O tenant deve ser propagado. | `x-tenant` vira `Event.OrgID` e é enviado como header ao Ledger Config. | `handler.go::getHeaders`, `request.go::ToEntity`, `ledger_config.go::FindConfigByLevel` |
| FR-04 | A configuração deve ser buscada por `processing_code` e `program_id`. | Chamada `GET {LEDGER_CONFIG_API_URL}/v1/ledger/config/{processing_code}/{program_id}` com headers `x-cid` e `x-tenant`. | `ledger_config.go::FindConfigByLevel` |
| FR-05 | O processador deve ser selecionado pelo campo `producer`. | `producer = regular` executa `ProcessRegular`; `producer = migration` executa `ProcessMigration`. | `services.go::p` |
| FR-06 | As entradas devem ser calculadas a partir dos scripts da config. | Para cada script com `flow == producer`, avalia `expression` com `Amounts` e `Fees`; cria uma `Entry` com `EntryTypeID = script_id`. | `regular.go`, `migration.go`, `response.go::RetrieveEntryByProducer` |
| FR-07 | O evento só é válido se possuir ao menos uma entrada. | `Event.Validate()` retorna erro quando `len(Entries) == 0`. | `domain/events/event.go::Validate` |
| FR-08 | Evento e entradas devem ser persistidos atomicamente. | `SaveEvent` inicia transação, insere em `event` e `entry` e faz rollback em caso de erro. | `mysql/repository.go::SaveEvent` |
| FR-09 | O evento deve ser publicado após a persistência. | `Topic.Emit` publica no tópico `EVENTS_SNS_TOPIC` com envelope `event_id`, `event_type = "ledger"`, `event_data` e `event_date`. | `services.go::CreateEvent`, `message/service.go::Emit` |
| FR-10 | Sucesso retorna `201` com o evento calculado. | Resposta contém `event_id`, `org_id`, `processing_code`, entradas etc., conforme `EventResponse`. | `handler.go::CreateEvent`, `rest/ledger/response.go` |
| FR-11 | Falhas de negócio e de integração retornam `400` com `{"message": "..."}`. | Erros de config, expressão, validação, persistência ou publicação são convertidos em `shared.ErrResponse`. | `handler.go::CreateEvent`, `rest/shared/erros.go` |
| FR-12 | O serviço deve expor health checks. | `GET /health/live` responde `200`; `GET /health/ready` responde `200` se MySQL e SNS estiverem saudáveis, `503` caso contrário. | `server.go::Liveness/Readiness` |
| FR-13 | O desligamento deve ser gracioso. | Ao receber `SIGINT`/`SIGTERM`, encerra o HTTP com timeout de 25s e depois o OTel. | `cmd/main.go` |
| FR-14 | A aplicação deve inicializar observabilidade. | `otel.Start(ctx)` no início e `otel.Shutdown(ctx)` no fim do `main`. | `cmd/main.go` |

## Requisitos de integração

| ID | Requisito | Critério de aceitação | Rastreabilidade |
|---|---|---|---|
| IR-01 | Consulta ao Ledger Config usa cliente HTTP instrumentado. | `httpclient.Client()` do core-sdk, com headers e tratamento de `404`. | `ledger_config.go::New` |
| IR-02 | Ledger Config possui circuit breaker. | `gobreaker` com nome `LedgerConfigAPI`, timeout de `15s` e abertura após mais de 5 falhas consecutivas. | `ledger_config.go::New` |
| IR-03 | Publicação usa o publisher SNS do core-sdk. | `sns.NewPublisher(ctx, config.New())` com `TopicARN` vindo de `EVENTS_SNS_TOPIC`. | `message/service.go::New` |
| IR-04 | Health de prontidão valida dependências externas. | `mysql.Repository.Health()` (ping) e `message.Topic.Health()` (`GetTopicAttributes`). | `server.go::Readiness` |

## Requisitos não funcionais

| ID | Requisito | Critério de aceitação | Rastreabilidade |
|---|---|---|---|
| NFR-01 | Rastreabilidade distribuída. | Spans nomeados `Service::CreateEvent`, `LedgerConfigApi::FindConfigByLevel`, `Repository::SaveEvent` e `Topic::Emit`, com atributo `cid`. | `internal/application/events`, `internal/infra` |
| NFR-02 | Logs estruturados. | Uso do logger `zap` do core-sdk e middlewares de log/tracing no Echo. | `server.go::routes`, `cmd/main.go` |
| NFR-03 | Configuração por variáveis de ambiente. | `config.New()` lê env via core-sdk; `APP_PORT` tem default `5001`. | `config/config.go` |
| NFR-04 | Testabilidade por inversão de dependência. | `Repository`, `ConfigProvider` e `Topic` são interfaces com mocks gerados por `mockgen`. | `interfaces.go`, `mock.go` |
| NFR-05 | Testes sem infraestrutura externa. | `go test ./...` executa com mocks e `httptest`. | `services_test.go`, `ledger_config_test.go` |
| NFR-06 | Aritmética decimal exata. | Valores usam `shopspring/decimal`; coluna `amount` é `decimal(20,15)`. | `domain/events/entry.go`, `scripts/sql/init.sql` |
| NFR-07 | Distribuição containerizada. | Dockerfile multi-stage (Go + Alpine) e manifests Kubernetes em `scripts/k8s/`. | `scripts/docker/api/Dockerfile`, `scripts/k8s/` |
| NFR-08 | Análise de performance. | Benchmarks e perfis disponíveis via targets `bench-*` e `profile-*` do Makefile. | `Makefile`, `bench_test.go` |

## Observações sobre requisitos

- `INFERIDO`: o campo `producer` deve ser `regular` ou `migration`. O código seleciona o processador por um mapa; um valor desconhecido não é validado explicitamente (ver `OBS-05`).
- `INFERIDO`: o contrato espera `x-tenant`, mas o handler apenas lê o header e não rejeita requisições sem ele.
- Erros de integração são mapeados para `400`, inclusive falhas de banco ou SNS; não há mapeamento para `5xx` no handler (ver `OBS-06`).
