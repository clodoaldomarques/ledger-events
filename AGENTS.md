# AGENTS.md — ledger-events

Single Go binary (`cmd/main.go`) with Echo routes in `internal/adapters/inbound/http/server`. The only business route is `POST /v1/ledger/events`; it expects `x-tenant` and generates `x-cid` when absent. The handler (`internal/adapters/inbound/http/ledger`) receives an injected `ports.EventUseCase`, binds the request to a DTO, and delegates to `internal/core/ledger/application.Service.CreateEvent`, which picks a domain processor by `producer` (`regular`/`migration`), saves the event, and emits it. `cmd/main.go` is the composition root: it builds the MySQL, Ledger Config, and SNS adapters once and injects them.

## Verification
- `go test ./...` passes with no Docker or external services: tests use gomock, in-memory fakes, and `httptest`; there is no integration suite. Focused runs: `go test -run TestService_CreateEvent ./internal/core/ledger/application`, `go test -run TestAccountconfigsApi_FindScriptByLevel ./internal/adapters/outbound/config`.
- `make arch-check` fails if `internal/core/**` imports adapters or infrastructure (AWS SDK, Echo, root `config`, MySQL driver, gobreaker).
- Avoid `make test` for quick checks: it writes `cover.out` and opens an HTML coverage viewer. No linter, CI workflow, or typecheck config is checked in; use `gofmt -l .` and `go vet ./...`.
- Benchmarks are in `internal/core/ledger/application/bench_test.go`. `make bench`, `make bench-one name=BenchmarkCreateEvent_Success`, `make bench-parallel`, and `make profile-*` already exclude tests with `-run=^$`; `bench-compare` requires `benchstat` installed separately.

## Config and local runtime
- `make run` exports `.env` before `go run cmd/main.go`; plain `go run ./cmd/...` does not load `.env` and leaves DB/AWS/API settings empty (`APP_PORT` alone defaults to 5001).
- `config.New()` is a `sync.Once` singleton: the first call fixes env-derived values, and later `With...` options mutate the shared instance. The config-adapter test now injects the httptest URL directly, but the singleton behavior remains for other callers. Do not expect fresh config per call.
- `docker compose up` starts dependencies only — there is no app service in compose. Local runs need MySQL `localhost:3306/ledger` (schema from `scripts/sql/init.sql`), LocalStack `:4566` with topic `events-sns-topic`, MockServer `:1080` (Ledger Config expectations in `scripts/docker/mockserver/expectations.json`), and OTLP `:4317`. `/health/ready` checks MySQL + SNS through an injected `ports.HealthChecker`; `/health/live` is dependency-free.
- Ledger Config expressions use core-sdk `expr-lang` env names `Amounts`, `Fees` (plus `Sum`, `Add`, `Sub`, `Mul`, `Div`), e.g. `Amounts.amount + Fees.iof`. Singular `Amount`/`Fee` will not resolve.

## Generated and stale files
- `internal/core/ledger/ports/mock.go` is generated: after editing `ports/outbound.go`, run `go generate ./internal/core/ledger/ports` (`mockgen`, `go.uber.org/mock`). Never edit the generated file by hand.
- `docs/collections/*.postman_collection.json` is stale: it targets `/v1/accounting/events` with old fields. Trust `internal/adapters/inbound/http/ledger/request.go`: the live payload uses `processing_code`, `program_id`, `account_id`, `producer`, `amounts`, `fees`.

## Infra targets
- `make up` runs `docker compose up -d` then `make terraform`; the terraform target blocks on a hardcoded `nc -z 192.168.67.2 30002` LocalStack wait before `plan`/`apply`. `make apply` deploys `scripts/k8s/` (expects `mysql-credentials` and `aws-credentials` secrets); `make destroy`/`reload` tear down.
