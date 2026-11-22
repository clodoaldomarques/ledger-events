# Tarefas: `ledger-events` — Adoção da Arquitetura Hexagonal

**Repo**: `github.com/clodoaldomarques/ledger-events` | **Ordem**: participante da família ledger/balances (spec canônica em `ledger-config`)

**Branch**: `005-hexagonal-adapters`

**Entrada**: [spec.md](spec.md) e [plan.md](plan.md) deste diretório; `docs/specs/constitution.md`, `docs/specs/platform.md` e specs `001`-`004` deste repo.

**Convenção**: `[ ]` = a fazer; `[x]` = concluído **no código** (não marcar apenas por ter escrito a spec). Cada checkpoint exige `go test ./...`, `go build ./...` e `go vet ./...` verdes.

**IDs**: `T001`-`T031` são os marcos comuns da família, adaptados ao escopo deste repo; `LE-001`-`LE-028` são os passos concretos que os implementam aqui. Os IDs são preservados para rastreabilidade, ainda que um marco e seus passos apareçam juntos.

## Fase 0: Baseline e guarda

- [x] T001 Rodar `go test ./...`, `go build ./...` e `go vet ./...` neste repo e registrar o resultado antes de qualquer mudança
- [x] T002 [P] Criar `scripts/check-architecture.sh`: falha se `internal/core/**` importar `internal/adapters`, AWS SDK, Echo, `config`, MySQL driver ou gobreaker
- [x] T003 [P] Adicionar alvo `arch-check` ao `Makefile` e documentá-lo em `docs/specs/platform.md` e no README
- [x] T004 Confirmar thread safety de `sns.Publisher`, `sql.DB` e do `httpclient` do core-sdk para uso compartilhado entre requests

**Checkpoint**: suíte verde e guarda arquitetural passando no estado atual.

---

## Fase 1: Movimento físico

- [x] T005 [P] Executar o mapeamento de pastas de [plan.md](plan.md) neste repo, sem alterar regras de negócio
- [x] T006 [P] Ajustar todos os imports e a diretiva `go:generate` (que vai com `interfaces.go` para `ports/`)
- [x] T007 Rodar `go generate ./...`, `go test ./...` e `go build ./...` após o movimento
- [x] LE-001 `git mv internal/domain/events → internal/core/ledger/domain`
- [x] LE-002 `git mv internal/application/events/services.go → internal/core/ledger/application/services.go`; mover `services_test.go` e `bench_test.go` junto
- [x] LE-003 `git mv internal/application/events/interfaces.go → internal/core/ledger/ports/outbound.go` (com a diretiva `go:generate`); criar `inbound.go` com `EventUseCase`
- [x] LE-004 `git mv internal/application/events/mock.go → internal/core/ledger/ports/mock.go`; regenerar, nunca editar à mão
- [x] LE-005 `git mv internal/infra/rest/ledger → internal/adapters/inbound/http/ledger`; mover `server` e `shared` para `internal/adapters/inbound/http/{server,shared}`
- [x] LE-006 `git mv internal/infra/db/mysql → internal/adapters/outbound/mysql`
- [x] LE-007 `git mv internal/infra/ledger/config → internal/adapters/outbound/config`
- [x] LE-008 `git mv internal/infra/message → internal/adapters/outbound/sns`
- [x] LE-009 Ajustar imports em `services_test.go`, `bench_test.go` e `ledger_config_test.go`; rodar a suíte completa

**Não aplicável aqui**: T008 (reparo do build de `balances-worker`) — este repo já compila.

**Checkpoint**: árvore alvo em disco, suíte verde, nenhuma regra alterada.

---

## Fase 2: Portas explícitas

- [x] T009 [P] Renomear portas de saída para nomes intencionais (`EventRepository`, `EventConfigProvider`, `EventPublisher`, `HealthChecker`)
- [x] T010 [P] Declarar a porta de entrada `EventUseCase` no core
- [x] T011 [P] Mudar construtores dos adapters para receber cliente/config em vez de chamar `config.New()` internamente
- [x] T012 [P] Regenerar mocks ao lado das portas e ajustar `services_test.go`/`bench_test.go`
- [x] T013 [US1] Criar teste de contrato da porta com fake in-memory executando a mesma suíte do caso de uso
- [x] LE-010 Renomear `Repository` → `EventRepository`, `ConfigProvider` → `EventConfigProvider`, `Topic` → `EventPublisher`
- [x] LE-011 Declarar `EventUseCase` com o método público `CreateEvent` hoje em `Service`
- [x] LE-012 Trocar `mysql.NewRepository(ctx)` por `mysql.New(db)`; mover `connection.go` e a obtenção da string de conexão para a raiz de composição
- [x] LE-013 Trocar `configApi.New(ctx)` por `config.New(client)`; configurar circuit breaker `LedgerConfigAPI` na raiz
- [x] LE-014 Trocar `message.New(ctx)` por `sns.New(publisher)`; manter envelope e tags atuais
- [x] LE-015 Definir `HealthChecker` no core e implementá-lo nos três adapters (ping MySQL, atributos do tópico SNS, GET de config quando aplicável)
- [x] LE-016 Criar teste de contrato das portas com fake in-memory (mesmos casos de `services_test.go`)

**Checkpoint**: US1 e US2 verificáveis; core sem import de adapter.

---

## Fase 3: Composition root e adaptadores de entrada

- [x] T014 [P] Mover a construção de adapters para `cmd/main.go` neste repo
- [x] T015 [US2][P] Transformar `CreateEvent` em método de struct que recebe a porta de entrada; remover `configApi.New`/`mysql.NewRepository`/`message.New` do handler
- [x] T016 [US3][P] Injetar `HealthChecker` em `/health/ready`
- [x] T017 [US3] Centralizar o mapeador de erros de domínio para status HTTP (400) no adaptador de entrada
- [x] T018 [US3] Garantir que `Close()` dos adapters seja chamado no graceful shutdown
- [x] T019 [US2] Adicionar teste de handler com fake da porta de entrada
- [x] LE-017 Em `cmd/main.go`: construir `config.New()`, cliente MySQL, `httpclient` do core-sdk, `sns.NewPublisher`, adapters, `Service`, handler e servidor
- [x] LE-018 Transformar `CreateEvent` de função solta em método de `Handler{uc EventUseCase}`; remover as três chamadas de construtores do handler
- [x] LE-019 Alterar `server.New()` para receber o handler e o `HealthChecker`; remover imports de `mysql`/`message` do servidor
- [x] LE-020 Mover `getHeaders` (x-cid/x-tenant) e `NewSpanFromContext` para `adapters/inbound/http/shared` (`NewSpanFromContext` não existe neste repo; spans seguem no core)
- [x] LE-021 Criar `adapters/inbound/http/shared/erros.go` com o mapeador `400` usado hoje no handler
- [x] LE-022 Fechar adapters no graceful shutdown em `cmd/main.go`; preservar o comportamento dos `Close()` atuais
- [x] LE-023 Adicionar teste de handler com fake de `EventUseCase` (sem MySQL, SNS ou MockServer)

**Checkpoint**: SC-003 e SC-005; contrato de `openapi.yaml` inalterado.

---

## Fase 4: Remoção de código morto

- [ ] T020 [US5] Migrar o teste do cliente de config de `github.com/golang/mock` para `go.uber.org/mock/gomock` neste repo
- [ ] T021 [US5] Remover `ErrDuplicatedEntry` e `ErrDuplicatedAccount` por nunca serem usados no fluxo
- [ ] T022 [US5] Rodar `go mod tidy` neste repo
- [ ] T023 [US5] Buscar `golang/mock`, `ErrDuplicatedEntry` e `ErrDuplicatedAccount` neste repo e atualizar `AGENTS.md`, `docs/specs/platform.md` e documentação
- [ ] LE-024 Migrar `adapters/outbound/config/*_test.go` para `go.uber.org/mock/gomock` mantendo os mesmos casos de teste
- [ ] LE-025 Excluir `ErrDuplicatedEntry` e `ErrDuplicatedAccount` de `internal/core/ledger/domain/errors.go`; manter `ErrScriptNotFound` e `ErrEntryNotFound`
- [ ] LE-026 Rodar `go mod tidy` e conferir `go.mod`/`go.sum`

**Checkpoint**: SC-006 satisfeito.

---

## Fase 5: Opcional / diferido (fora dos critérios de sucesso)

- [ ] T024 [P] Avaliar decorators para mover `tracer`/`logger` do core para as bordas
- [ ] T025 [P] Avaliar divisão de portas grossas por capacidade (ex.: `EventReader`/`EventWriter`)
- [ ] T026 Corrigir desvios `OBS-01`..`OBS-11` das specs 001-004 em specs próprias, com coordenação de contrato

---

## Fase 6: Documentação e fechamento

- [ ] T027 [P] Reescrever o Princípio I da constituição deste repo ("regra de dependência via portas") e atualizar versão/data
- [ ] T028 [P] Atualizar `docs/specs/README.md`, `platform.md` e os planos `001`-`004` com os novos caminhos
- [ ] T029 [P] Registrar a convenção arquitetural neste repo
- [ ] T030 Rodar `go test ./...`, `go build ./...`, `go vet ./...`, `make arch-check`
- [ ] T031 Comparar respostas de `POST /v1/ledger/events` e `/health/*` com `docs/specs/openapi.yaml` e conferir SC-001 a SC-007
- [ ] LE-027 Reescrever Princípio I em `docs/specs/constitution.md`: "regra de dependência via portas"; atualizar versão e data; ajustar índice e planos
- [ ] LE-028 Rodar verificação final: `go test ./...`, `go build ./...`, `go vet ./...`, `make arch-check`

## Dependências e Ordem de Execução

- Fase 0 bloqueia todas as demais.
- Fase 1 deve estar verde antes da Fase 2.
- Fase 2 bloqueia Fase 3; Fase 4 pode ser paralela à Fase 3 neste repo.
- A spec canônica em `ledger-config` define a convenção da família; mudanças nos IDs `T001`-`T031` exigem alinhamento prévio.
- Fase 5 depende de decisão explícita e não é necessária para fechar a spec.
