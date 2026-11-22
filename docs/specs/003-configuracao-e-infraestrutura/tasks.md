# Tarefas: Configuração, Infraestrutura e Verificação

**Entrada**: documentos em `/specs/003-configuracao-e-infraestrutura/`

**Pré-requisitos**: `plan.md` (obrigatório), `spec.md` (obrigatório), `../platform.md`

**Convenção**: `[x]` = já implementado em `396db3e`; `[ ]` = lacuna identificada.

**Formato**: `[ID] [P?] [US] Descrição com caminho exato do arquivo`

## Fase 1: Setup (Infraestrutura Comum)

**Objetivo**: ambiente de desenvolvimento declarado.

- [x] T001 [P] Criar `docker-compose.yaml` com MySQL, LocalStack, MockServer, OTel, Zipkin, Prometheus, Redis e Stackport
- [x] T002 [P] Criar `scripts/sql/init.sql` com o DDL de `event` e `entry`
- [x] T003 [P] Criar `scripts/docker/mockserver/expectations.json` com as expectativas da config
- [x] T004 [P] Criar `scripts/terraform/{provider,sns}.tf` para o tópico `events-sns-topic`

**Checkpoint**: `docker compose up -d` sobe as dependências.

---

## Fase 2: US1 - Configurar a aplicação por ambiente (Prioridade: P1) 🎯 MVP

**Objetivo**: singleton de configuração e defaults.

**Teste independente**: `make run` com `.env` → porta `5001`.

- [x] T005 [US1] Implementar `config.New()` com `sync.Once` em `config/config.go`
- [x] T006 [US1] Implementar `WithAppPort`, `WithAwsAddress`, `WithAwsRegion` e `WithLedgerConfigApiUrl` em `config/config.go`
- [x] T007 [US1] Definir default `5001` para `APP_PORT` em `config/config.go`
- [x] T008 [US1] Implementar alvo `run` que exporta `.env` antes de `go run cmd/main.go` no `Makefile`
- [ ] T009 [US1] Separar construção de config de sobrescritas para evitar vazamento entre testes (`OBS-10`)

**Checkpoint**: aplicação configurável sem recompilação.

---

## Fase 3: US2 - Subir as dependências locais (Prioridade: P1)

**Objetivo**: MySQL, SNS, MockServer e observabilidade disponíveis.

**Teste independente**: `curl http://localhost:1080` e verificação do tópico no LocalStack.

- [x] T010 [US2] Documentar portas e Papéis dos serviços em `../platform.md`
- [x] T011 [US2] Implementar alvo `up` (compose + terraform) e `down` no `Makefile`
- [x] T012 [US2] Implementar alvo `terraform` com espera `nc -z 192.168.67.2 30002` no `Makefile`
- [ ] T013 [US2] Parametrizar o endpoint do LocalStack em Terraform e no `Makefile` (`OBS-09`)

**Checkpoint**: dependências locais operacionais.

---

## Fase 4: US3 - Verificar e medir o serviço (Prioridade: P2)

**Objetivo**: suíte determinística, benchmarks e perfis.

**Teste independente**: `go test ./...` e `make bench-one name=BenchmarkCreateEvent_Success`.

- [x] T014 [US3] Implementar `services_test.go` e `ledger_config_test.go` sem serviços externos
- [x] T015 [US3] Implementar `bench_test.go` com `BenchmarkCreateEvent_Success` e `BenchmarkCreateEvent_Parallel`
- [x] T016 [US3] Implementar alvos `test`, `bench-*`, `profile-*` e `profile-clean` no `Makefile`
- [x] T017 [US3] Documentar `go test ./...`, `gofmt -l .` e `go vet ./...` em `../platform.md`
- [ ] T018 [US3] Documentar a instalação de `benchstat` ou remover o alvo `bench-compare`

**Checkpoint**: verificação reproduzível.

---

## Fase 5: US4 - Construir, publicar e aplicar em Kubernetes (Prioridade: P2)

**Objetivo**: entrega containerizada.

**Teste independente**: `make build version=1.0.0` e inspeção das tags.

- [x] T019 [US4] Criar `scripts/docker/api/Dockerfile` multi-stage
- [x] T020 [US4] Implementar alvos `build`, `push`, `publish` e `version` no `Makefile`
- [x] T021 [US4] Criar `scripts/k8s/` com Deployment, Service `LoadBalancer` (5001/nodePort 31001) e ConfigMap `ledger-schema`
- [x] T022 [US4] Implementar alvos `apply`, `destroy` e `reload` no `Makefile`
- [ ] T023 [US4] Gerar o ConfigMap `ledger-schema` a partir de `scripts/sql/init.sql` (`OBS-07`)
- [ ] T024 [US4] Atualizar a collection Postman para `POST /v1/ledger/events` (`OBS-08`)

**Checkpoint**: deploy reproduzível; lacunas permanecem como exercício.

---

## Fase N: Polimento

- [ ] T025 [P] Revisar `README.md` e `AGENTS.md` após mudanças de comandos
- [ ] T026 Rodar `go test ./...`, `gofmt -l .`, `go vet ./...` e `go build ./...`

## Dependências e Ordem de Execução

- Fase 1 bloqueia US2.
- US1 é independente das demais após a Fase 1.
- US3 depende da suíte implementada nas specs 001 e 002.
- US4 depende de US2 (Terraform) e deve ser validada com as secrets do cluster.
- Correções T009, T013, T018, T023 e T024 viram specs próprias se alterarem contrato externo.
