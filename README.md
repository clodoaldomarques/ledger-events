# Ledger Events

> Go backend service responsible for managing accounting events within the Ledger platform.

## Overview

Ledger Events receives accounting-event requests, resolves the processing configuration from Ledger Config, calculates entries from amounts and fees, persists the event in MySQL, and publishes it to SNS for asynchronous processing by Ledger Worker.

It is part of a broader set of Go backend services:

```text
                          ┌───────────────────┐
                          │   Ledger Config   │
                          │   Configuration   │
                          └─────────┬─────────┘
                                    │
                          ┌─────────▼─────────┐
                          │   Ledger Events   │
                          │       Go API      │
                          └─────────┬─────────┘
                                    │ Events
                          ┌─────────▼─────────┐
                          │   Ledger Worker   │
                          │ Async Processing  │
                          └───────────────────┘
```

Shared infrastructure (AWS clients, OTel, logging, expression evaluation) comes from [`core-sdk`](https://github.com/clodoaldomarques/core-sdk).

## Architecture

Request flow:

1. `cmd/main.go` starts OpenTelemetry and the Echo HTTP server, then handles graceful shutdown.
2. `internal/infra/rest/server` wires middlewares (tracing, logging, OTel), health routes, and the events route.
3. `internal/infra/rest/ledger` binds the request and builds the MySQL, SNS, and Ledger Config clients for that request.
4. `internal/application/events.Service.CreateEvent` looks up the config, chooses a processor by `producer`, validates, saves, and emits the event.
5. `internal/domain/events` holds the pure rules; expressions are evaluated with `core-sdk/pkg/expression` against the variables `Amounts` and `Fees` (plus `Sum`, `Add`, `Sub`, `Mul`, `Div`).
6. `internal/infra` contains the adapters: MySQL repository, SNS publisher, Ledger Config HTTP client with circuit breaker, and REST handlers.
7. `config` exposes an environment-backed `Config` singleton used by all adapters.

## Project layout

```text
cmd/                  application entry point
config/               environment-backed configuration
docs/                 C4 diagrams and Postman collection
internal/
  application/events  orchestration service, interfaces, generated mocks
  domain/events       event model and processing rules
  infra/db/mysql      persistence
  infra/ledger/config Ledger Config HTTP client
  infra/message       SNS publisher
  infra/rest          Echo server, handlers, request/response types
scripts/
  docker/             Dockerfile, MockServer expectations, OTel/Prometheus config
  k8s/                deployment, service, schema ConfigMap
  sql/                MySQL schema used by docker compose
  terraform/          SNS topic provisioned on LocalStack
```

## API

### `POST /v1/ledger/events`

Headers:

- `x-tenant`: tenant identifier, forwarded to Ledger Config and stored on the event.
- `x-cid`: correlation id; generated automatically when absent.

Request body:

```json
{
  "processing_code": "101",
  "program_id": 401,
  "account_id": 123,
  "producer": "migration",
  "amounts": { "amount": 150.00 },
  "fees": { "iof": 15.99 }
}
```

`producer` must be `regular` or `migration`. Ledger Config is queried at `GET {LEDGER_CONFIG_API_URL}/v1/ledger/config/{processing_code}/{program_id}`; its scripts express calculations such as `Amounts.amount + Fees.iof`. A successful call returns `201` with the saved event and calculated entries.

### Health

- `GET /health/live` — process is up, no dependency checks.
- `GET /health/ready` — checks MySQL and SNS availability.

## Requirements

- Go 1.26.1+ (see `go.mod`)
- Docker and Docker Compose
- Make
- Optional: Terraform and `nc` for `make up`/`make apply`, `kubectl` for deployment, `mockgen` for mock regeneration, `benchstat` for `make bench-compare`

## Configuration

Create a local `.env` file (gitignored). The current development values are:

| Variable | Default / example |
|---|---|
| `APP_PORT` | `5001` (default in code) |
| `MYSQL_HOST`, `MYSQL_PORT`, `MYSQL_DATABASE` | `localhost`, `3306`, `ledger` |
| `MYSQL_USER`, `MYSQL_PASSWORD` | `admin`, `adminpassword` |
| `AWS_ADDRESS` | `http://localhost:4566` (LocalStack) |
| `AWS_REGION` | `us-east-1` |
| `AWS_ACCESS_KEY_ID`, `AWS_SECRET_ACCESS_KEY` | `test`, `test` |
| `EVENTS_SNS_TOPIC` | `arn:aws:sns:us-east-1:000000000000:events-sns-topic` |
| `LEDGER_CONFIG_API_URL` | `http://localhost:1080` (MockServer) |
| `OTEL_EXPORTER_OTLP_ENDPOINT` | `localhost:4317` |
| `OTEL_SERVICE_NAME` | `ledger-events` |

`make run` exports `.env` before starting the app. Running `go run ./cmd/...` directly does **not** load it; only `APP_PORT` has a code default.

## Running locally

```bash
go mod download
docker compose up -d       # dependencies only: MySQL, LocalStack, MockServer, OTel, etc.
make run                   # exports .env and starts the API on :5001
```

There is no application container in `docker compose`: it starts only the services the API depends on. Useful local endpoints: MockServer `:1080`, LocalStack `:4566`, Zipkin `:9411`, Prometheus `:9090`, OTLP collector `:4317`/`:4318`.

`make up` also runs Terraform to create the SNS topic, but its provider and readiness check are hardcoded to LocalStack at `192.168.67.2:30002`. If LocalStack is only available on `localhost:4566`, use `docker compose up -d` plus `make run` instead.

Quick checks:

```bash
curl http://localhost:5001/health/live
curl http://localhost:5001/health/ready
```

Use `make down` to stop the dependencies and remove their volumes.

## Tests and benchmarks

```bash
go test ./...                                   # no external services required
make arch-check                                 # fails if internal/core imports adapters or infra
make test                                       # coverage in cover.out + HTML viewer
go test -run TestService_CreateEvent ./internal/application/events
go test -run TestAccountconfigsApi_FindScriptByLevel ./internal/infra/ledger/config
```

Benchmarks live in `internal/application/events/bench_test.go`:

```bash
make bench
make bench-one name=BenchmarkCreateEvent_Success
make bench-parallel
make profile-cpu
```

`make bench-compare` needs `benchstat` installed separately. After changing `internal/application/events/interfaces.go`, regenerate mocks with `go generate ./internal/application/events`.

## Deployment

```bash
make build version=1.0.0                        # builds and tags the container image
make publish version=1.0.0                      # build + push
make apply                                      # terraform + kubectl apply -f scripts/k8s/
make destroy                                    # remove k8s and terraform-managed resources
```

Kubernetes deployment expects the `mysql-credentials` and `aws-credentials` secrets, and applies the schema from `scripts/k8s/configmap.yaml` via an init container. The service exposes port `5001` (node port `31001`). Terraform state is local and gitignored.

## Documentation and notes

- C4 diagrams (Mermaid): `docs/model/*.mmd`.
- Reverse-engineered specs in Spec Kit format (`spec.md`/`plan.md`/`tasks.md`): [`docs/specs/`](docs/specs/README.md).
- The Postman collection in `docs/collections/` is stale: it targets `/v1/accounting/events` with an older payload. Use `POST /v1/ledger/events` and the request shape above.
- `AGENTS.md` contains repo-specific guidance for AI coding agents.

## Project status

This is a Go backend engineering study and portfolio project, not a production-ready financial platform.

## Author

**Clodoaldo Marques** — Backend Software Engineer focused on Go, Microservices, Distributed Systems and Cloud-Native architectures.

- GitHub: https://github.com/clodoaldomarques
- LinkedIn: https://www.linkedin.com/in/clodoaldomarques/
