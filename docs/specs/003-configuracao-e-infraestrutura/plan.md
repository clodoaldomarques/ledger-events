# Plano de Implementação: Configuração, Infraestrutura e Verificação

**Branch**: `003-configuracao-e-infraestrutura` | **Data**: 2026-11-22 | **Spec**: [spec.md](spec.md)

**Entrada**: especificação derivada de `config/`, `Makefile`, `docker-compose.yaml`, `scripts/`, `.env` e `internal/application/events/bench_test.go`.

**Nota**: este plano descreve decisões **já presentes** no código; não é um plano a executar do zero. A tabela completa de variáveis e serviços está em [`../platform.md`](../platform.md).

## Resumo

Configuração por ambiente com singleton `sync.Once`, dependências locais via Docker Compose (sem aplicação no Compose), build multi-stage, Terraform para o tópico SNS, manifests Kubernetes e uma suíte de verificação baseada em `go test`, benchmarks e perfis. Nenhuma ferramenta de CI/lint está configurada.

## Contexto Técnico

**Linguagem/Versão**: Go 1.26.1 (`go.mod`)

**Dependências principais**: core-sdk v0.8.8, Echo v4.16.0, MySQL driver, AWS SDK v2, OpenTelemetry 1.47

**Armazenamento**: MySQL local `localhost:3306/ledger`; schema em `scripts/sql/init.sql` e ConfigMap `ledger-schema`

**Testes**: `go test ./...` com mocks e `httptest`; `make test` grava `cover.out` e abre HTML

**Benchmarks**: `BenchmarkCreateEvent_Success`, `BenchmarkCreateEvent_Parallel` e perfis em `internal/application/events/bench_test.go`

**Plataforma alvo**: container Linux (Dockerfile em `scripts/docker/api/`), Kubernetes via `scripts/k8s/`

**Tipo de projeto**: serviço HTTP com infraestrutura declarada

**Metas de desempenho**: não declaradas; ferramentas de medição disponíveis via Makefile

**Restrições**: `.env` não é carregado pelo binário; Terraform depende de IP fixo; `benchstat` externo

**Escopo**: 1 singleton de config, 8 serviços no Compose, ~20 alvos de Makefile, Terraform + k8s

### Mapa de comandos

```text
make run                  # exporta .env e executa cmd/main.go
make up                   # docker compose up -d + make terraform
make down                 # docker compose down -v
make terraform            # init + espera nc + plan + apply (192.168.67.2:30002)
make apply / destroy      # kubectl apply/delete + terraform
make build / push / publish
make test                 # cover.out + viewer HTML
make bench / bench-count / bench-one / bench-baseline / bench-compare / bench-parallel
make profile-cpu / profile-mem / profile-mem-inuse / profile-goroutine / profile-trace / profile-clean
```

Todos os alvos `bench-*` e `profile-*` usam `-run=^$` para excluir testes.

## Verificação da Constituição

*GATE: aplicado antes de Phase 0 e reavaliado após o desenho.*

| Princípio | Situação | Evidência |
|---|---|---|
| I. Separação em camadas | Passa | `config` isolado; adapters em `internal/infra` |
| II. O domínio calcula, a interface transporta | N/A | Configuração e deploy não contêm regras de domínio |
| III. Configuração singleton | Passa com ressalva | `sync.Once` + options mutáveis (`OBS-10`) |
| IV. Observabilidade no request | Passa | Variáveis OTLP e interceptadores registrados no servidor |
| V. Persistência transacional antes da publicação | N/A | Fora do escopo desta spec |
| VI. Testes unitários determinísticos | Passa | `go test ./...` sem dependências externas |
| VII. Infraestrutura como código | Passa com ressalva | Compose/Terraform/k8s/Dockerfile; endpoint fixo (`OBS-09`) |

## Estrutura do Projeto

### Documentação (esta funcionalidade)

```text
docs/specs/003-configuracao-e-infraestrutura/
├── spec.md
├── plan.md
└── tasks.md
```

### Código-fonte e infraestrutura (raiz do repositório)

```text
.env                                      # gitignored, exportado por make run
config/config.go                          # singleton sync.Once + With...
Makefile                                  # alvos de execução, teste, bench, profile, deploy
docker-compose.yaml                       # dependências locais
scripts/docker/api/Dockerfile             # imagem multi-stage (Go + Alpine)
scripts/docker/mockserver/expectations.json
scripts/docker/otel-collector.yaml        # configuração OTLP
scripts/sql/init.sql                      # DDL canônico
scripts/terraform/{provider,sns}.tf       # tópico SNS no LocalStack
scripts/k8s/{deployment,service,configmap}.yaml
internal/application/events/bench_test.go
```

**Decisão de estrutura**: o binário não carrega `.env`; a responsabilidade é do Makefile. O schema tem duas cópias (`OBS-07`) e o endpoint do LocalStack é fixo (`OBS-09`).

## Complexidade Conhecida

| Desvio | Por que existe | Alternativa mais simples considerada |
|---|---|---|
| `OBS-07`: DDL duplicado em SQL e ConfigMap | Facilitar local e k8s independentemente | Gerar o ConfigMap a partir do SQL canônico |
| `OBS-08`: collection Postman desatualizada | Contrato antigo `/v1/accounting/events` | Atualizar collection e environment para `POST /v1/ledger/events` |
| `OBS-09`: endpoint fixo `192.168.67.2:30002` | Ambiente de desenvolvimento original | Parametrizar LocalStack em Terraform e no Makefile |
| `OBS-10`: `sync.Once` com `With...` mutáveis | Facilitar injeção em testes | Separar construção de config de sobrescritas |
| `benchstat` fora do módulo | Ferramenta de comparação opcional | Documentar instalação ou remover o alvo |
| `make test` abre navegador | `go tool cover -html` | Usar `go test -coverprofile` em checagens não interativas |
