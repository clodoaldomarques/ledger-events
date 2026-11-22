# Constituição — Ledger Events

> Derivada do baseline `396db3e` e reescrita após a adoção da topologia hexagonal pela spec `005-hexagonal-adapters`. Descreve o estado atual do projeto.

## Princípios Fundamentais

### I. Regra de dependência via portas (NÃO NEGOCIÁVEL)

O core (`internal/core/ledger/**`) não conhece adapters, AWS SDK, Echo, config de ambiente, driver MySQL nem gobreaker. As portas ficam em `internal/core/ledger/ports`: `EventUseCase` na entrada e `EventRepository`, `EventConfigProvider`, `EventPublisher` e `HealthChecker` na saída. A dependência aponta sempre do adapter para o core.

`cmd/main.go` é o único lugar autorizado a conhecer adapters concretos e configuração: constrói cada adapter uma vez, injeta o caso de uso no handler HTTP e publica o `HealthChecker` no servidor. Handlers apenas transportam dados; o domínio não conhece HTTP, SQL, AWS ou Echo. `make arch-check` falha se um import proibido aparecer no core.

### II. O domínio calcula, a interface transporta

`Event.Validate()`, `ProcessRegular` e `ProcessMigration` concentram as regras de negócio; expressões são avaliadas com `Amounts`, `Fees`, `Sum`, `Add`, `Sub`, `Mul` e `Div`. `EventRequest`/`ConfigResponse` apenas convertem dados externos em entidades. As tags `validate:"required"` existem, mas o handler chama `evt.Validate()` (que retorna `nil`) em vez do validador do Echo — lacuna registrada como `OBS-04`.

### III. Configuração por ambiente, com singleton explícito

`config.New()` usa `sync.Once`: as variáveis de ambiente são capturadas na primeira chamada. O binário não lê `.env` sozinho; `make run` exporta o arquivo antes de executar. As opções `With...` alteram a instância compartilhada (`OBS-10`). A raiz de composição entrega a configuração aos adapters; nenhum segredo é versionado.

### IV. Observabilidade é parte do request

Todo request passa pelos interceptadores de tracer, logger e instrumentação OTel do `core-sdk`. `x-cid` identifica a correlação (gerado quando ausente em `adapters/inbound/http/shared`) e `x-tenant` identifica a organização. Os spans `Service::CreateEvent`, `LedgerConfigApi::FindConfigByLevel`, `Repository::SaveEvent` e `Topic::Emit` carregam o atributo `cid` e permanecem no core/adapters nesta fase.

### V. Persistência transacional antes da publicação

`SaveEvent` grava `event` e `entry` em uma única transação, com rollback em caso de erro. A publicação no SNS ocorre depois do commit, sem outbox, retry ou idempotência — uma falha posterior deixa o evento persistido e não publicado (`OBS-11`). Erros de integração são mapeados para `400` (`OBS-06`).

### VI. Testes unitários determinísticos

A suíte usa `go test ./...`, `testify`, mocks `gomock` gerados por `mockgen` a partir de `internal/core/ledger/ports/outbound.go`, fakes in-memory das portas e `httptest` nos handlers. Não há testes de integração com MySQL/SNS na suíte padrão. Benchmarks e perfis ficam em `internal/core/ledger/application/bench_test.go` e são acionados pelos alvos `bench-*`/`profile-*` do Makefile.

### VII. Infraestrutura como código, com ambientes locais explícitos

`docker-compose.yaml`, `scripts/terraform/`, `scripts/k8s/` e `scripts/docker/` descrevem o ambiente. O Compose sobe apenas dependências (não há serviço de aplicação nele). `make up` depende de um endereço fixo do LocalStack (`192.168.67.2:30002`) no provider Terraform e na espera com `nc` (`OBS-09`).

## Restrições e Convenções

- Go 1.26.1 conforme `go.mod`; imagem de deploy baseada em Go + Alpine.
- Adapters são construídos na raiz e compartilhados entre requests; handlers nunca chamam construtores de adapter.
- Erros de caso de uso são convertidos em `400` com corpo `{"message": "..."}` pelo mapeador do adaptador de entrada; health checks não têm corpo.
- O contrato vivo é `POST /v1/ledger/events`; a collection em `docs/collections/` está desatualizada (`OBS-08`).
- Expressões usam os nomes `Amounts`/`Fees` (não `Amount`/`Fee`) e decimais `shopspring/decimal`.
- Nomes de benchmark e perfil devem ser conferidos em `internal/core/ledger/application/bench_test.go` antes de usar o Makefile.
- `mock.go` é gerado: altere `ports/outbound.go` e rode `go generate ./internal/core/ledger/ports`.
- `make arch-check` faz parte da verificação obrigatória antes do merge.

## Fluxo de Desenvolvimento

1. Especificar a mudança em `docs/specs/` (spec, plan e tasks da funcionalidade afetada).
2. Ajustar entidades e regras em `internal/core/ledger/domain/` quando a regra mudar.
3. Atualizar `internal/core/ledger/ports/outbound.go` e rodar `go generate ./internal/core/ledger/ports` se uma porta mudar.
4. Implementar caso de uso, handler e DTOs; manter `openapi.yaml` alinhado se o contrato HTTP mudar.
5. Rodar `go test ./...` e `make arch-check`; `make test` apenas quando desejar abrir o relatório HTML.
6. Verificar `gofmt -l .`, `go vet ./...` e `go build ./...` (não há CI, lint ou typecheck configurados).

## Governança

Esta constituição registra o estado atual. Alterações de arquitetura ou de regras de domínio devem atualizar este arquivo, a spec afetada e o `plan.md` correspondente na mesma mudança. A topologia hexagonal foi adotada pela spec `005-hexagonal-adapters/`; novas mudanças estruturais exigem especificação própria. Complexidade adicional precisa ser justificada na seção *Complexidade Conhecida* do plano.

**Versão**: 2.0 | **Ratificada**: 2026-11-21 | **Última alteração**: 2026-11-22
