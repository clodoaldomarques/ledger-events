# Tarefas: Criação e Publicação de Eventos Contábeis

**Entrada**: documentos em `/specs/001-criacao-de-eventos/`

**Pré-requisitos**: `plan.md` (obrigatório), `spec.md` (obrigatório), `../platform.md`

**Convenção**: `[x]` = já implementado em `396db3e`; `[ ]` = lacuna identificada na engenharia reversa.

**Formato**: `[ID] [P?] [US] Descrição com caminho exato do arquivo`

## Fase 1: Setup (Infraestrutura Comum)

**Objetivo**: projeto Go executável com ambiente local disponível.

- [x] T001 Criar módulo `github.com/clodoaldomarques/ledger-events` em `go.mod`
- [x] T002 [P] Declarar dependências Echo, core-sdk, MySQL driver, gobreaker, decimal, uuid e testify em `go.mod`
- [x] T003 [P] Definir `.env` com MySQL, AWS/LocalStack, SNS, Ledger Config e variáveis OTEL
- [x] T004 [P] Criar `scripts/sql/init.sql`, `scripts/docker/mockserver/expectations.json` e `scripts/terraform/sns.tf`

**Checkpoint**: `go mod download` e `docker compose up -d` disponíveis.

---

## Fase 2: Fundacional (Pré-requisitos Bloqueantes)

**Objetivo**: camadas compartilhadas por todas as histórias.

- [x] T005 [P] Implementar singleton de ambiente em `config/config.go`
- [x] T006 [P] Implementar entidades `Event`/`Entry` e `Validate()` em `internal/core/ledger/domain/{event,entry}.go`
- [x] T007 [P] Implementar interfaces `Config`/`Script` e DTOs em `internal/core/ledger/domain/configs.go` e `internal/adapters/outbound/config/response.go`
- [x] T008 [P] Implementar `ProcessRegular`/`ProcessMigration` em `internal/core/ledger/domain/{processing,regular,migration}.go`
- [x] T009 [P] Declarar portas `Repository`, `ConfigProvider` e `Topic` em `internal/core/ledger/ports/outbound.go`
- [x] T010 [P] Gerar mocks com `go generate ./internal/core/ledger/ports` em `internal/core/ledger/ports/mock.go`
- [x] T011 [P] Registrar rota e interceptadores em `internal/adapters/inbound/http/server/server.go`
- [x] T012 [P] Implementar DTOs de request/response em `internal/adapters/inbound/http/ledger/{request,response}.go`
- [x] T013 Implementar cliente HTTP com circuit breaker em `internal/adapters/outbound/config/ledger_config.go`
- [x] T014 Implementar repositório MySQL e mapeamento em `internal/adapters/outbound/mysql/{repository,type,connection}.go`
- [x] T015 Implementar publicador SNS e envelope em `internal/adapters/outbound/sns/{service,types}.go`

**Checkpoint**: base pronta; as histórias podem ser verificadas isoladamente.

---

## Fase 3: US1 - Receber e correlacionar a requisição (Prioridade: P1) 🎯 MVP

**Objetivo**: vincular o corpo, ler headers e gerar correlation id.

**Teste independente**: `POST /v1/ledger/events` com `x-tenant` e sem `x-cid` → UUID propagado.

- [x] T016 [US1] Implementar `getHeaders` e a criação dos clients por request em `internal/adapters/inbound/http/ledger/handler.go`
- [x] T017 [US1] Implementar `EventRequest.ToEntity` com `uuid.NewString()` e `time.Now()` em `internal/adapters/inbound/http/ledger/request.go`
- [x] T018 [US1] Implementar bind, `evt.Validate()` e mapeamento de erro para `400` em `internal/adapters/inbound/http/ledger/handler.go`
- [x] T019 [US1] Construir `EventResponse`/`EntryResponse` em `internal/adapters/inbound/http/ledger/response.go`
- [ ] T020 [US1] Aplicar as tags `validate:"required"` chamando `c.Validate(evt)` em `internal/adapters/inbound/http/ledger/handler.go` (`OBS-04`)

**Checkpoint**: recebimento e resposta funcionando de forma independente.

---

## Fase 4: US2 - Resolver a configuração no Ledger Config (Prioridade: P1)

**Objetivo**: consultar a config instrumentada e tratá-la como porta de domínio.

**Teste independente**: cliente apontado para `httptest` com expectativa `/v1/ledger/config/101/401`.

- [x] T021 [US2] Implementar `FindConfigByLevel` em `internal/adapters/outbound/config/ledger_config.go`
- [x] T022 [US2] Implementar circuit breaker `LedgerConfigAPI` com 15s e limite de 5 falhas em `internal/adapters/outbound/config/ledger_config.go`
- [x] T023 [US2] Implementar parse e mensagens de erro (`404`, status inesperado, unmarshal) em `internal/adapters/outbound/config/{ledger_config,errors}.go`
- [x] T024 [US2] Cobrir sucesso, `404` e `unmarshal` em `internal/adapters/outbound/config/ledger_config_test.go`
- [ ] T025 [US2] Adicionar teste de integração contra o MockServer local para a expectativa real de config

**Checkpoint**: consulta utilizável isoladamente.

---

## Fase 5: US3 - Calcular as entradas com as regras de domínio (Prioridade: P1)

**Objetivo**: despachar o produtor, avaliar expressões e validar o evento.

**Teste independente**: `amounts.amount = 150`, `fees.iof = 15.99`, expressão `Amounts.amount + Fees.iof` → entrada `165.99`.

- [x] T026 [US3] Implementar `Service.CreateEvent` com despacho por `producer` em `internal/core/ledger/application/services.go`
- [x] T027 [US3] Implementar cálculo com `expression` e decimais em `internal/core/ledger/domain/{regular,migration,processing}.go`
- [x] T028 [US3] Implementar `Event.Validate()` e `ErrEntryNotFound` em `internal/core/ledger/domain/{event,errors}.go`
- [x] T029 [US3] Cobrir sucesso, config ausente, expressão inválida e evento sem entradas em `internal/core/ledger/application/services_test.go`
- [ ] T030 [US3] Validar `producer` contra `regular`/`migration` antes do despacho em `internal/core/ledger/application/services.go` (`OBS-05`)

**Checkpoint**: cálculo validado de forma independente.

---

## Fase 6: US4 - Persistir evento e entradas em uma transação (Prioridade: P1)

**Objetivo**: gravar evento e entradas atomicamente.

**Teste independente**: com MySQL local, conferir uma linha em `event` e N em `entry` com o mesmo `event_id`.

- [x] T031 [US4] Implementar `buildEventTable`/`buildEntryTable` em `internal/adapters/outbound/mysql/type.go`
- [x] T032 [US4] Implementar `SaveEvent` com `BeginTx`, inserts e `defer tx.Rollback()` em `internal/adapters/outbound/mysql/repository.go`
- [x] T033 [US4] Implementar `Health()` (ping) e `Close()` em `internal/adapters/outbound/mysql/repository.go`
- [ ] T034 [US4] Propagar erro de conexão em `mysql.NewRepository` em vez de devolver `db == nil` (`OBS-02`)
- [ ] T035 [US4] Verificar erro de `buildEventTable` e retorno de `tx.Commit()` em `internal/adapters/outbound/mysql/repository.go` (`OBS-03`)

**Checkpoint**: persistência validada; lacunas marcadas permanecem como exercício.

---

## Fase 7: US5 - Publicar o evento e responder ao cliente (Prioridade: P1)

**Objetivo**: fechar o caso de uso com a publicação e o `201`.

**Teste independente**: com mocks, verificar ordem `SaveEvent` → `Emit` e envelope com `event_type = "ledger"`.

- [x] T036 [US5] Implementar `Topic.Emit` e `ToEventMessage` em `internal/adapters/outbound/sns/{service,types}.go`
- [x] T037 [US5] Publicar após a persistência em `internal/core/ledger/application/services.go`
- [x] T038 [US5] Implementar `Health()`/`Close()` do publicador em `internal/adapters/outbound/sns/service.go`
- [x] T039 [US5] Documentar o contrato em `docs/specs/openapi.yaml`
- [ ] T040 [US5] Corrigir `make([]Entry, len(et))` para `make([]Entry, 0, len(et))` e cobrir payload em `internal/adapters/outbound/sns/types.go` (`OBS-01`)
- [ ] T041 [US5] Diferenciar erros de validação (`4xx`) de erros de dependência (`5xx`) no handler (`OBS-06`)
- [ ] T042 [US5] Avaliar outbox transacional ou retry/idempotência para a publicação (`OBS-11`)

**Checkpoint**: caso de uso completo; contrato congelado em `openapi.yaml`.

---

## Fase N: Polimento

- [ ] T043 [P] Revisar `README.md`, `AGENTS.md` e `docs/specs/` após correções
- [ ] T044 [P] Atualizar a collection em `docs/collections/` para `POST /v1/ledger/events` (`OBS-08`)
- [ ] T045 [P] Adicionar testes de contrato HTTP com `httptest`
- [ ] T046 Rodar `go test ./...`, `gofmt -l .`, `go vet ./...` e `go build ./...`

## Dependências e Ordem de Execução

- Fase 2 bloqueia todas as histórias.
- US1 a US5 são independentes entre si após a Fase 2, mas a verificação de US5 depende de US3 e US4.
- Correções T020, T025, T030, T034-T035 e T040-T042 devem ser feitas após a story correspondente, com testes de regressão.
- Mudanças de contrato HTTP exigem atualização de `spec.md`, `plan.md` e `openapi.yaml` na mesma alteração.
