# Constituição — Ledger Events

> Derivada da implementação no baseline `396db3e`. Esta é uma constituição **descritiva**: registra o que o projeto já faz hoje.

## Princípios Fundamentais

### I. Separação em camadas (NÃO NEGOCIÁVEL)

O fluxo é `cmd/main.go` → `internal/infra/rest/server` → `internal/infra/rest/ledger` (handler e DTOs) → `internal/application/events` (caso de uso) → `internal/domain/events` (regras). Os adapters de infraestrutura ficam em `internal/infra/{db/mysql,ledger/config,message,rest}`. Handlers não calculam entradas; o domínio não conhece HTTP, SQL, AWS ou Echo. As portas hoje vivem em `internal/application/events/interfaces.go`.

### II. O domínio calcula, a interface transporta

`Event.Validate()`, `ProcessRegular` e `ProcessMigration` concentram as regras de negócio; expressões são avaliadas com `Amounts`, `Fees`, `Sum`, `Add`, `Sub`, `Mul` e `Div`. `EventRequest`/`ConfigResponse` apenas convertem dados externos em entidades. As tags `validate:"required"` existem, mas o handler chama `evt.Validate()` (que retorna `nil`) em vez do validador do Echo — lacuna registrada como `OBS-04`.

### III. Configuração por ambiente, com singleton explícito

`config.New()` usa `sync.Once`: as variáveis de ambiente são capturadas na primeira chamada. O binário não lê `.env` sozinho; `make run` exporta o arquivo antes de executar. As opções `With...` alteram a instância compartilhada — é assim que o teste do cliente de config injeta a URL do `httptest` (`OBS-10`). Nenhum segredo é versionado.

### IV. Observabilidade é parte do request

Todo request passa pelos interceptadores de tracer, logger e instrumentação OTel do `core-sdk`. `x-cid` identifica a correlação (gerado quando ausente) e `x-tenant` identifica a organização. Os spans `Service::CreateEvent`, `LedgerConfigApi::FindConfigByLevel`, `Repository::SaveEvent` e `Topic::Emit` carregam o atributo `cid`.

### V. Persistência transacional antes da publicação

`SaveEvent` grava `event` e `entry` em uma única transação, com rollback em caso de erro. A publicação no SNS ocorre depois do commit, sem outbox, retry ou idempotência — uma falha posterior deixa o evento persistido e não publicado (`OBS-11`). Erros de integração são mapeados para `400` (`OBS-06`).

### VI. Testes unitários determinísticos

A suíte usa `go test ./...`, `testify` e mocks `gomock` gerados por `mockgen` a partir de `internal/application/events/interfaces.go`; o cliente de config usa `httptest`. Não há testes de integração com MySQL/SNS na suíte padrão. Benchmarks e perfis ficam em `internal/application/events/bench_test.go` e são acionados pelos alvos `bench-*`/`profile-*` do Makefile.

### VII. Infraestrutura como código, com ambientes locais explícitos

`docker-compose.yaml`, `scripts/terraform/`, `scripts/k8s/` e `scripts/docker/` descrevem o ambiente. O Compose sobe apenas dependências (não há serviço de aplicação nele). `make up` depende de um endereço fixo do LocalStack (`192.168.67.2:30002`) no provider Terraform e na espera com `nc` (`OBS-09`).

## Restrições e Convenções

- Go 1.26.1 conforme `go.mod`; imagem de deploy baseada em Go + Alpine.
- Handler cria cliente de config, repositório MySQL e publicador SNS **por requisição** e os fecha com `defer`.
- Erros de caso de uso são convertidos em `400` com corpo `{"message": "..."}`; health checks não têm corpo.
- O contrato vivo é `POST /v1/ledger/events`; a collection em `docs/collections/` está desatualizada (`OBS-08`).
- Expressões usam os nomes `Amounts`/`Fees` (não `Amount`/`Fee`) e decimais `shopspring/decimal`.
- Nomes de benchmark e perfil devem ser conferidos em `internal/application/events/bench_test.go` antes de usar o Makefile.
- `mock.go` é gerado: altere `interfaces.go` e rode `go generate ./internal/application/events`.

## Fluxo de Desenvolvimento

1. Especificar a mudança em `docs/specs/` (spec, plan e tasks da funcionalidade afetada).
2. Ajustar entidades e regras em `internal/domain/events/` quando a regra mudar.
3. Atualizar `interfaces.go` e rodar `go generate ./internal/application/events` se uma porta mudar.
4. Implementar caso de uso, handler e DTOs; manter `openapi.yaml` alinhado se o contrato HTTP mudar.
5. Rodar `go test ./...`; `make test` apenas quando desejar abrir o relatório HTML.
6. Verificar `gofmt -l .`, `go vet ./...` e `go build ./...` (não há CI, lint ou typecheck configurados).

## Governança

Esta constituição é um registro do estado atual. Alterações de arquitetura ou de regras de domínio devem atualizar este arquivo, a spec afetada e o `plan.md` correspondente na mesma mudança. A adoção da topologia hexagonal é acompanhada pela spec prescritiva `005-hexagonal-adapters/`, que prevê a reescrita do Princípio I e dos planos ao final da migração. Complexidade adicional precisa ser justificada na seção *Complexidade Conhecida* do plano.

**Versão**: 1.0 | **Ratificada**: 2026-11-21 | **Última alteração**: 2026-11-22
