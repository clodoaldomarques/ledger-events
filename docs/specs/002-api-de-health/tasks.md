# Tarefas: API de Health Checks

**Entrada**: documentos em `/specs/002-api-de-health/`

**Pré-requisitos**: `plan.md` (obrigatório), `spec.md` (obrigatório), `../platform.md`

**Convenção**: `[x]` = já implementado em `396db3e`; `[ ]` = lacuna identificada.

**Formato**: `[ID] [P?] [US] Descrição com caminho exato do arquivo`

## Fase 1: Setup (Rotas)

**Objetivo**: servidor Echo com as rotas de sonda registradas.

- [x] T001 [P] Criar `Server` e `New()` em `internal/adapters/inbound/http/server/server.go`
- [x] T002 [P] Registrar `GET /health/live` e `GET /health/ready` em `internal/adapters/inbound/http/server/server.go`
- [x] T003 Documentar ambas as rotas em `docs/specs/openapi.yaml`

**Checkpoint**: `curl http://localhost:5001/health/live` retorna `200`.

---

## Fase 2: US1 - Responder liveness (Prioridade: P1) 🎯 MVP

**Objetivo**: sonda independente de dependências.

**Teste independente**: derrubar MySQL e SNS e chamar `GET /health/live` → `200`.

- [x] T004 [US1] Implementar `Server.Liveness` com `c.NoContent(http.StatusOK)` em `internal/adapters/inbound/http/server/server.go`

**Checkpoint**: liveness verificada isoladamente.

---

## Fase 3: US2 - Verificar dependências na prontidão (Prioridade: P1)

**Objetivo**: checar MySQL e SNS por chamada.

**Teste independente**: com dependências no ar, `GET /health/ready` → `200`; sem MySQL, → `503`.

- [x] T005 [US2] Implementar contexto com timeout de `500ms` em `Server.Readiness`
- [x] T006 [US2] Criar `mysql.NewRepository(ctx)`, chamar `Health()` e fazer `defer Close()` em `internal/adapters/inbound/http/server/server.go`
- [x] T007 [US2] Implementar `Repository.Health()` com `PingContext` em `internal/adapters/outbound/mysql/repository.go`
- [x] T008 [US2] Criar `message.New(ctx)`, chamar `Health()` e implementar `Topic.Health()` via publisher em `internal/adapters/outbound/sns/service.go`
- [ ] T009 [US2] Adicionar teste com `httptest` e fakes de `Health()` cobrindo `200` e `503`

**Checkpoint**: prontidão validada com as duas dependências.

---

## Fase 4: US3 - Falhar fechado sem corpo (Prioridade: P2)

**Objetivo**: contrato de sonda previsível.

**Teste independente**: simular falha de SNS e conferir `503` com corpo vazio.

- [x] T010 [US3] Responder `c.NoContent(http.StatusServiceUnavailable)` em qualquer falha em `internal/adapters/inbound/http/server/server.go`
- [ ] T011 [US3] Propagar erro de conexão em `mysql.NewRepository` para evitar `db == nil` na prontidão (`OBS-02`)

**Checkpoint**: contrato de sonda congelado.

---

## Fase N: Polimento

- [ ] T012 [P] Revisar `openapi.yaml` após qualquer mudança de status das sondas
- [ ] T013 Rodar `go test ./...`, `gofmt -l .`, `go vet ./...` e `go build ./...`

## Dependências e Ordem de Execução

- Fase 1 bloqueia as histórias.
- US2 depende dos adapters de MySQL e SNS descritos na spec 001.
- T009 e T011 devem ser feitos com testes de regressão; T011 depende da decisão tomada em `001` (`OBS-02`).
