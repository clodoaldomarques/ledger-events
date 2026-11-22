# Tarefas: Observabilidade e Desligamento Gracioso

**Entrada**: documentos em `/specs/004-observabilidade-e-shutdown/`

**Pré-requisitos**: `plan.md` (obrigatório), `spec.md` (obrigatório), `../platform.md`

**Convenção**: `[x]` = já implementado em `396db3e`; `[ ]` = lacuna identificada.

**Formato**: `[ID] [P?] [US] Descrição com caminho exato do arquivo`

## Fase 1: Setup (Telemetria Básica)

**Objetivo**: coletor e dependências disponíveis no ambiente local.

- [x] T001 [P] Declarar OTel Collector, Zipkin e Prometheus em `docker-compose.yaml`
- [x] T002 [P] Configurar OTLP gRPC/HTTP e exportadores em `scripts/docker/otel-collector.yaml`
- [x] T003 [P] Declarar `OTEL_EXPORTER_OTLP_ENDPOINT` e `OTEL_SERVICE_NAME` no `.env`

**Checkpoint**: `docker compose up -d` expõe OTLP em `4317` e Zipkin em `9411`.

---

## Fase 2: US1 - Correlacionar requisições (Prioridade: P1) 🎯 MVP

**Objetivo**: `cid` presente em spans, logs e banco.

**Teste independente**: enviar `x-cid` conhecido e localizar o valor no span e em `event.correlation_id`.

- [x] T004 [US1] Implementar leitura/geração de `x-cid` em `internal/infra/rest/ledger/handler.go`
- [x] T005 [US1] Passar `cid` para `Service.CreateEvent` e para o span `Service::CreateEvent` em `internal/application/events/services.go`
- [x] T006 [US1] Propagar `cid` aos spans de config, repositório e publicador em `internal/infra/{ledger/config,db/mysql,message}`
- [x] T007 [US1] Gravar `cid` em `event.correlation_id` em `internal/infra/db/mysql/type.go`
- [ ] T008 [US1] Adicionar teste automatizado que afirme o atributo `cid` nos spans

**Checkpoint**: correlação verificada ponta a ponta.

---

## Fase 3: US2 - Emitir spans, logs e métricas (Prioridade: P1)

**Objetivo**: instrumentação HTTP e quatro spans nomeados.

**Teste independente**: requisição bem-sucedida visível no Zipkin com os quatro spans.

- [x] T009 [US2] Registrar `tracer.Interceptor()`, `logger.Interceptor()` e `otelecho.Interceptor()` em `internal/infra/rest/server/server.go`
- [x] T010 [US2] Criar span `Service::CreateEvent` em `internal/application/events/services.go`
- [x] T011 [US2] Criar span `LedgerConfigApi::FindConfigByLevel` em `internal/infra/ledger/config/ledger_config.go`
- [x] T012 [US2] Criar span `Repository::SaveEvent` em `internal/infra/db/mysql/repository.go`
- [x] T013 [US2] Criar span `Topic::Emit` em `internal/infra/message/service.go`
- [x] T014 [US2] Configurar exportação de métricas para Prometheus no coletor
- [ ] T015 [US2] Instalar tracer noop em `TestMain` para eliminar dependência de coletor nos testes

**Checkpoint**: telemetria observável no ambiente local.

---

## Fase 4: US3 - Iniciar e encerrar a telemetria com o processo (Prioridade: P2)

**Objetivo**: startup e shutdown ordenados.

**Teste independente**: `SIGTERM` → mensagens de shutdown e HTTP encerrado antes do OTel.

- [x] T016 [US3] Chamar `otel.Start(ctx)` no início de `cmd/main.go`
- [x] T017 [US3] Registrar `signal.Notify` para `SIGINT`/`SIGTERM` em `cmd/main.go`
- [x] T018 [US3] Criar `context.WithTimeout(ctx, 25*time.Second)` no shutdown em `cmd/main.go`
- [x] T019 [US3] Chamar `Server.Shutdown(ctx)` e depois `otel.Shutdown(ctx)` em `cmd/main.go`
- [x] T020 [US3] Imprimir mensagens/erros no stdout sem alterar o status de saída em `cmd/main.go`
- [ ] T021 [US3] Tornar o timeout de shutdown configurável por variável de ambiente
- [ ] T022 [US3] Propagar erro de shutdown como status de saída não zero

**Checkpoint**: ciclo de vida completo.

---

## Fase N: Polimento

- [ ] T023 [P] Revisar nomes de spans em `spec.md`, `plan.md` e `../platform.md`
- [ ] T024 Rodar `go test ./...`, `gofmt -l .`, `go vet ./...` e `go build ./...`

## Dependências e Ordem de Execução

- Fase 1 bloqueia US2.
- US1 depende do handler descrito na spec 001.
- US2 depende dos adapters instrumentados nas specs 001 e 002.
- US3 pode ser validada em paralelo a US2, mas a ordem HTTP → OTel é obrigatória.
- T008 e T015 devem usar tracer noop para manter a suíte determinística.
