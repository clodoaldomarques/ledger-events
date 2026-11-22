# Especificação da Funcionalidade: Configuração, Infraestrutura e Verificação

**Branch da funcionalidade**: `003-configuracao-e-infraestrutura`

**Criado**: 2026-11-22

**Status**: Rascunho (derivado do código em `396db3e`)

**Entrada**: engenharia reversa de `config/config.go`, `.env`, `Makefile`, `docker-compose.yaml`, `scripts/docker/`, `scripts/k8s/`, `scripts/terraform/` e `internal/application/events/bench_test.go`

> **Escopo**: espelho `AS-IS` do suporte de execução do serviço. Detalhes de stack e variáveis estão em [`../platform.md`](../platform.md).

## Visão Geral

O binário obtém configuração exclusivamente de variáveis de ambiente; `make run` exporta o `.env` antes de executar. O Docker Compose sobe apenas dependências, sem serviço de aplicação. Build, deploy e verificações são expostos por alvos do Makefile; benchmarks e perfis cobrem o caminho de criação de eventos.

## Cenários de Usuário e Testes *(obrigatório)*

### História de Usuário 1 - Configurar a aplicação por ambiente (Prioridade: P1)

Um desenvolvedor precisa apontar o serviço para MySQL, LocalStack, MockServer e OTLP sem recompilar.

**Por que esta prioridade**: nenhum valor de infraestrutura é embutido no binário; apenas `APP_PORT` tem default no código.

**Teste independente**: executar `make run` com o `.env` de desenvolvimento e confirmar que a aplicação escuta em `5001` e conecta nas dependências locais.

**Cenários de aceitação**:

1. **Dado** `APP_PORT` ausente, **Quando** `config.New()` é chamado, **Então** a porta é `5001`.
2. **Dado** `MYSQL_HOST/PORT/DATABASE/USER/PASSWORD` no `.env`, **Quando** o repositório é criado, **Então** a string de conexão usa `localhost:3306/ledger`.
3. **Dado** `AWS_ADDRESS=http://localhost:4566`, **Quando** o publisher SNS é criado, **Então** ele aponta para o LocalStack.
4. **Dado** `LEDGER_CONFIG_API_URL=http://localhost:1080`, **Quando** o cliente de config é criado, **Então** ele consulta o MockServer.
5. **Dado** uma chamada anterior a `config.New()`, **Quando** outra chamada ocorre, **Então** os valores de env já fixados são reaproveitados (`sync.Once`); `With...` muta a instância compartilhada (`OBS-10`).
6. **Dado** `go run ./cmd/...` sem exportar `.env`, **Quando** a aplicação inicia, **Então** DB/AWS/API ficam com valores vazios — apenas `APP_PORT` tem default.

---

### História de Usuário 2 - Subir as dependências locais (Prioridade: P1)

Um desenvolvedor precisa de MySQL, SNS, config mockada e observabilidade no ambiente local.

**Por que esta prioridade**: o repositório não contém serviço de aplicação no Compose; sem essas dependências o fluxo de eventos não pode ser exercitado.

**Teste independente**: `docker compose up -d` e verificar portas `3306`, `4566`, `1080`, `4317`, `9411`, `9090`.

**Cenários de aceitação**:

1. **Dado** `docker compose up -d`, **Quando** o Compose termina, **Então** MySQL, LocalStack, MockServer, OTel Collector, Zipkin, Prometheus, Redis e Stackport estão disponíveis conforme `platform.md`.
2. **Dado** o schema em `scripts/sql/init.sql`, **Quando** o MySQL sobe, **Então** as tabelas `event` e `entry` são criadas via `docker-entrypoint-initdb.d`.
3. **Dado** as expectativas em `scripts/docker/mockserver/expectations.json`, **Quando** o cliente de config consulta `/v1/ledger/config/101/401`, **Então** o MockServer responde com scripts e descrição.
4. **Dado** `make up`, **Quando** o Terraform é executado, **Então** o comando espera `192.168.67.2:30002` com `nc` e cria o tópico SNS nesse endpoint (`OBS-09`).
5. **Dado** apenas `localhost:4566` disponível, **Quando** `make up` é executado, **Então** ele falha na espera; use `docker compose up -d` + `make run`.

---

### História de Usuário 3 - Verificar e medir o serviço (Prioridade: P2)

Um mantenedor precisa executar testes, benchmarks e perfis sem depender de serviços externos.

**Por que esta prioridade**: a suíte é determinística por mocks e `httptest`, e os benchmarks instrumentam o caminho de criação de eventos.

**Teste independente**: `go test ./...`, `make bench-one name=BenchmarkCreateEvent_Success` e `make profile-cpu`.

**Cenários de aceitação**:

1. **Dado** o repositório clonado, **Quando** `go test ./...` é executado, **Então** todos os testes passam sem Docker ou serviços externos.
2. **Dado** o teste focado, **Quando** `go test -run TestService_CreateEvent ./internal/application/events` é executado, **Então** apenas o caso de uso é verificado.
3. **Dado** `gofmt -l .` e `go vet ./...`, **Quando** executados, **Então** não apontam problemas — não há linter, CI ou typecheck configurados.
4. **Dado** `make test`, **Quando** executado, **Então** grava `cover.out` e abre visualizador HTML; não é adequado para checagens não interativas.
5. **Dado** os alvos `bench-*`/`profile-*`, **Quando** executados, **Então** usam `-run=^$` para excluir testes.
6. **Dado** `bench-compare`, **Quando** executado sem `benchstat`, **Então** falha: a ferramenta não é dependência do módulo.
7. **Dado** uma mudança em `interfaces.go`, **Quando** `go generate ./internal/application/events` é executado, **Então** `mock.go` é regenerado.

---

### História de Usuário 4 - Construir, publicar e aplicar em Kubernetes (Prioridade: P2)

Um operador precisa gerar a imagem e aplicar os manifests.

**Por que esta prioridade**: fecha o ciclo de entrega do serviço.

**Teste independente**: `make build version=1.0.0` e conferir as tags `1.0.0` e `latest`.

**Cenários de aceitação**:

1. **Dado** `make build version=1.0.0`, **Quando** o build termina, **Então** existem `clodoaldomarques/ledger-events:1.0.0` e `:latest`, gerados por `scripts/docker/api/Dockerfile`.
2. **Dado** `make publish`, **Quando** executado, **Então** ambas as tags são enviadas.
3. **Dado** `make apply`, **Quando** executado, **Então** Terraform cria o tópico SNS e `kubectl apply -f scripts/k8s/` aplica Deployment, Service e ConfigMap.
4. **Dado** o Deployment, **Quando** ele inicia, **Então** usa as secrets `mysql-credentials` e `aws-credentials`, e um init container aplica o schema do ConfigMap `ledger-schema`.
5. **Dado** o Service, **Quando** exposto, **Então** é `LoadBalancer` na porta `5001` com `nodePort: 31001`.
6. **Dado** `make destroy`, **Quando** executado, **Então** remove manifests e recursos do Terraform; o state é local e gitignored.

---

### Casos Limite

- DDL duplicado entre `scripts/sql/init.sql` e `scripts/k8s/configmap.yaml` pode divergir (`OBS-07`).
- A collection em `docs/collections/` usa `/v1/accounting/events` e campos antigos (`OBS-08`).
- Terraform e Makefile fixam `192.168.67.2:30002` (`OBS-09`).
- `config.New()` é `sync.Once`; options mutam o singleton e podem vazar estado entre testes (`OBS-10`).
- Redis está no Compose, mas não é usado pelo código atual.
- `make up` depende de `nc` e Terraform instalados; `make apply` depende de `kubectl`.

## Requisitos *(obrigatório)*

### Requisitos Funcionais

- **FR-001**: O sistema DEVE obter toda a configuração de variáveis de ambiente, com `APP_PORT` default `5001`. *(NFR-03)*
- **FR-002**: O sistema DEVE disponibilizar `make run` que exporta o `.env` antes de executar `cmd/main.go`.
- **FR-003**: O sistema DEVE expor `config.New()` como singleton `sync.Once`, com opções `With...` explícitas. *(OBS-10)*
- **FR-004**: O sistema DEVE declarar em Compose apenas as dependências necessárias ao desenvolvimento local, sem serviço de aplicação.
- **FR-005**: O sistema DEVE aplicar o schema em `scripts/sql/init.sql` no MySQL local e no init container do Kubernetes. *(OBS-07)*
- **FR-006**: O sistema DEVE oferecer alvos de teste, benchmark e perfilagem que não dependam de serviços externos. *(NFR-05, NFR-08)*
- **FR-007**: O sistema DEVE gerar mocks via `go generate ./internal/application/events` a partir de `interfaces.go`. *(NFR-04)*
- **FR-008**: O sistema DEVE construir imagem multi-stage e publicar tags `<version>` e `latest`. *(NFR-07)*
- **FR-009**: O sistema DEVE aplicar manifests Kubernetes com as secrets `mysql-credentials` e `aws-credentials`.
- **FR-010**: O sistema DEVE manter a collection Postman alinhada ao contrato vivo `POST /v1/ledger/events`. *(OBS-08)*

### Entidades Principais

- **Config**: singleton em `config/config.go`, com getters para MySQL, AWS, SNS, Ledger Config e porta HTTP.
- **Compose**: serviços MySQL, LocalStack, MockServer, OTel Collector, Zipkin, Prometheus, Redis e Stackport.
- **Makefile**: alvos `up`, `down`, `run`, `build`, `push`, `publish`, `terraform`, `apply`, `destroy`, `test`, `bench-*`, `profile-*`.
- **Manifests**: `scripts/k8s/` (Deployment, Service, ConfigMap) e `scripts/terraform/` (provider, SNS).

## Critérios de Sucesso *(obrigatório)*

### Resultados Mensuráveis

- **SC-001**: `go test ./...` passa sem Docker nem serviços externos.
- **SC-002**: `make run` inicia o serviço na porta `5001` usando apenas variáveis de ambiente.
- **SC-003**: `docker compose up -d` expõe todas as dependências listadas em `platform.md`.
- **SC-004**: `make build version=<tag>` produz as tags `<tag>` e `latest`.
- **SC-005**: `make apply` aplica o schema e o Deployment com as secrets esperadas.
- **SC-006**: `make bench`, `make bench-parallel` e `make profile-*` executam com `-run=^$`.

## Suposições

- `.env` existe localmente e é gitignored; nenhum segredo é versionado.
- Docker, Compose, Make, Terraform, `nc` e `kubectl` estão disponíveis conforme a necessidade de cada alvo.
- `benchstat` é instalado separadamente quando se usa `make bench-compare`.
- A ausência de CI é intencional neste projeto de estudo.
