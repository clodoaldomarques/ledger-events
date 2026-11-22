# 06 — Configuração, operação e dívidas conhecidas

## Variáveis de ambiente

Fonte: `config/config.go` e `.env` (gitignored). `make run` exporta o `.env` antes de executar a aplicação.

| Variável | Default no código | Valor de desenvolvimento |
|---|---|---|
| `APP_PORT` | `5001` | não definido no `.env` |
| `MYSQL_HOST` | `""` | `localhost` |
| `MYSQL_PORT` | `""` | `3306` |
| `MYSQL_DATABASE` | `""` | `ledger` |
| `MYSQL_USER` | `""` | `admin` |
| `MYSQL_PASSWORD` | `""` | `adminpassword` |
| `AWS_ADDRESS` | `""` | `http://localhost:4566` |
| `AWS_REGION` | `""` | `us-east-1` |
| `AWS_ACCESS_KEY_ID` | `""` | `test` |
| `AWS_SECRET_ACCESS_KEY` | `""` | `test` |
| `EVENTS_SNS_TOPIC` | `""` | `arn:aws:sns:us-east-1:000000000000:events-sns-topic` |
| `LEDGER_CONFIG_API_URL` | `""` | `http://localhost:1080` |
| `OTEL_EXPORTER_OTLP_ENDPOINT` | — | `localhost:4317` |
| `OTEL_SERVICE_NAME` | — | `ledger-events` |

`config.New()` usa `sync.Once`: a primeira chamada fixa os valores lidos do ambiente. As opções `WithAppPort`, `WithAwsAddress`, `WithAwsRegion` e `WithLedgerConfigApiUrl` alteram a instância compartilhada — é assim que o teste do cliente de config injeta a URL do `httptest` (`OBS-10`).

## Serviços locais

`docker compose up -d` sobe apenas dependências; não existe serviço de aplicação no compose.

| Serviço | Porta | Função |
|---|---|---|
| MySQL | `3306` | banco `ledger`, schema em `scripts/sql/init.sql` |
| LocalStack | `4566` | SNS, SQS e DynamoDB |
| MockServer | `1080` | respostas do Ledger Config em `scripts/docker/mockserver/expectations.json` |
| OTel Collector | `4317`, `4318`, `8888` | recepção OTLP e métricas para Prometheus |
| Zipkin | `9411` | visualização de traços |
| Prometheus | `9090` | scrapa `otel-collector:8888/metrics` |
| Redis | `6379` | disponível no ambiente, sem uso no código atual |
| Stackport | `8080` | painel local de recursos AWS |

## Comandos

```bash
# desenvolvimento
go mod download
docker compose up -d
make run

# verificação
go test ./...
go vet ./...
gofmt -l .
make test                         # gera cover.out e abre viewer HTML

# foco e performance
go test -run TestService_CreateEvent ./internal/application/events
make bench
make bench-one name=BenchmarkCreateEvent_Success
make bench-parallel
make profile-cpu

# build e deploy
make build version=1.0.0
make publish version=1.0.0
make apply
make destroy
make down
```

`bench-compare` depende de `benchstat`, que não faz parte das dependências do módulo. `make up` executa `make terraform`, que espera LocalStack em `192.168.67.2:30002` e usa esse mesmo endereço no provider AWS (`OBS-09`).

## Observabilidade

- Middlewares no Echo: tracer, logger e instrumentação OTel do core-sdk.
- Spans com atributo `cid`: `Service::CreateEvent`, `LedgerConfigApi::FindConfigByLevel`, `Repository::SaveEvent`, `Topic::Emit`.
- `otel.Start(ctx)` no startup e `otel.Shutdown(ctx)` no graceful shutdown.
- O coletor envia traços para Zipkin; logs e métricas são exportados em modo `debug`; o Prometheus scrapa as métricas do coletor.

## Deploy

- `make build` gera a imagem `clodoaldomarques/ledger-events:<version>` e a tag `latest` usando `scripts/docker/api/Dockerfile`.
- `make apply` roda Terraform (tópico SNS) e aplica `scripts/k8s/`.
- O Deployment usa as secrets `mysql-credentials` e `aws-credentials`, e um init container aplica o schema do ConfigMap `ledger-schema`.
- O Service é do tipo `LoadBalancer` na porta `5001` (`nodePort: 31001`).
- O state do Terraform é local e gitignored.

## Dívidas e desvios observados

Estes itens são `AS-IS`: descrevem o código atual, não uma intenção de projeto. Em um fluxo SDD, cada item pode virar um novo requisito.

| ID | Evidência | Impacto | Encaminhamento sugerido |
|---|---|---|---|
| `OBS-01` | `message/types.go::ToEntryMessage` usa `make([]Entry, len(et))` e depois `append`. | A mensagem SNS contém N entradas vazias seguidas das N entradas reais. | Construir o slice com capacidade (`make([]Entry, 0, len(et))`) e cobrir com teste de payload. |
| `OBS-02` | `mysql/repository.go::NewRepository` retorna `&Repository{}` quando `Connect` falha. | `SaveEvent` e `Close` podem operar com `db == nil`. | Retornar erro ou tornar explícito o estado inválido; tratar no handler. |
| `OBS-03` | `SaveEvent` ignora o erro de `buildEventTable` e o retorno de `tx.Commit()`. | Falhas de escrita podem passar despercebidas. | Verificar ambos os erros e fechar statements preparados. |
| `OBS-04` | `server.go` define `Validator`, mas o handler chama `evt.Validate()`, que sempre retorna `nil`. | As tags `validate:"required"` de `EventRequest` não são aplicadas. | Chamar `c.Validate(evt)` ou validar explicitamente os campos obrigatórios. |
| `OBS-05` | `services.go` usa `p[e.Producer]` sem verificar a existência da chave. | Um `producer` desconhecido causa chamada de função nil. | Validar `producer` contra `regular`/`migration` antes de despachar. |
| `OBS-06` | `handler.go` responde `400` para qualquer erro do serviço. | Falhas de MySQL/SNS são reportadas como erro do cliente, dificultando observação e retries. | Diferenciar erros de validação de erros de dependência (`4xx` vs `5xx`). |
| `OBS-07` | DDL duplicado em `scripts/sql/init.sql` e `scripts/k8s/configmap.yaml`. | Mudanças de schema podem divergir entre local e Kubernetes. | Gerar o ConfigMap a partir do arquivo SQL canônico. |
| `OBS-08` | `docs/collections/Accounting-Events.postman_collection.json` usa `/v1/accounting/events` e campos antigos. | Quem usa a collection testa um contrato que não existe. | Atualizar collection e environment para `POST /v1/ledger/events`. |
| `OBS-09` | `scripts/terraform/provider.tf` e `Makefile::terraform` fixam `192.168.67.2:30002`. | `make up`/`make apply` falham se o LocalStack não estiver nesse endereço. | Parametrizar o endpoint do LocalStack. |
| `OBS-10` | `config.New()` é `sync.Once` com options que mutam o singleton. | Estado global pode vazar entre testes ou components. | Separar construção de config de sobrescritas, ou reiniciar em testes. |
| `OBS-11` | `SaveEvent` e `Topic.Emit` ocorrem em etapas separadas, sem outbox/retry. | Evento pode ficar persistido sem publicação. | Avaliar outbox transacional ou idempotência + retry na publicação. |
