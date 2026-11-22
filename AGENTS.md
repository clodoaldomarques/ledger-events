# AGENTS.md — ledger-events

Single Go binary (`cmd/main.go`) with Echo routes in `internal/infra/rest/server`. The only business route is `POST /v1/ledger/events`; it expects `x-tenant` and generates `x-cid` when absent. The handler (`internal/infra/rest/ledger`) builds MySQL, SNS, and Ledger Config clients per request, then `internal/application/events.Service.CreateEvent` picks a domain processor by `producer` (`regular`/`migration`), saves the event, and emits it.

## Verification
- `go test ./...` passes with no Docker or external services: tests use gomock and `httptest`, there is no integration suite. Focused runs: `go test -run TestService_CreateEvent ./internal/application/events`, `go test -run TestAccountconfigsApi_FindScriptByLevel ./internal/infra/ledger/config`.
- Avoid `make test` for quick checks: it writes `cover.out` and opens an HTML coverage viewer. No linter, CI workflow, or typecheck config is checked in; use `gofmt -l .` and `go vet ./...`.
- Benchmarks are in `internal/application/events/bench_test.go`. `make bench`, `make bench-one name=BenchmarkCreateEvent_Success`, `make bench-parallel`, and `make profile-*` already exclude tests with `-run=^$`; `bench-compare` requires `benchstat` installed separately.

## Config and local runtime
- `make run` exports `.env` before `go run cmd/main.go`; plain `go run ./cmd/...` does not load `.env` and leaves DB/AWS/API settings empty (`APP_PORT` alone defaults to 5001).
- `config.New()` is a `sync.Once` singleton: the first call fixes env-derived values, and later `With...` options mutate the shared instance (tests inject the httptest URL this way). Do not expect fresh config per call.
- `docker compose up` starts dependencies only — there is no app service in compose. Local runs need MySQL `localhost:3306/ledger` (schema from `scripts/sql/init.sql`), LocalStack `:4566` with topic `events-sns-topic`, MockServer `:1080` (Ledger Config expectations in `scripts/docker/mockserver/expectations.json`), and OTLP `:4317`. `/health/ready` checks MySQL + SNS; `/health/live` is dependency-free.
- Ledger Config expressions use core-sdk `expr-lang` env names `Amounts`, `Fees` (plus `Sum`, `Add`, `Sub`, `Mul`, `Div`), e.g. `Amounts.amount + Fees.iof`. Singular `Amount`/`Fee` will not resolve.

## Generated and stale files
- `internal/application/events/mock.go` is generated: after editing `interfaces.go`, run `go generate ./internal/application/events` (`mockgen`, `go.uber.org/mock`). The config test still imports legacy `github.com/golang/mock/gomock`; both modules are dependencies.
- `docs/collections/*.postman_collection.json` is stale: it targets `/v1/accounting/events` with old fields. Trust `internal/infra/rest/ledger/request.go`: the live payload uses `processing_code`, `program_id`, `account_id`, `producer`, `amounts`, `fees`.

## Infra targets
- `make up` runs `docker compose up -d` then `make terraform`; the terraform target blocks on a hardcoded `nc -z 192.168.67.2 30002` LocalStack wait before `plan`/`apply`. `make apply` deploys `scripts/k8s/` (expects `mysql-credentials` and `aws-credentials` secrets); `make destroy`/`reload` tear down.
