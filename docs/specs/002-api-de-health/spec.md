# Especificação da Funcionalidade: API de Health Checks

**Branch da funcionalidade**: `002-api-de-health`

**Criado**: 2026-11-22

**Status**: Rascunho (derivado do código em `396db3e`)

**Entrada**: engenharia reversa de `internal/infra/rest/server/server.go`, `internal/infra/db/mysql/repository.go`, `internal/infra/message/service.go`, `cmd/main.go` e `docs/specs/openapi.yaml`

> **Escopo**: espelho `AS-IS` das rotas `GET /health/live` e `GET /health/ready`. O contrato legível por máquina está em [`../openapi.yaml`](../openapi.yaml).

## Visão Geral

O servidor Echo expõe duas rotas de sonda. `/health/live` responde ao processo sem consultar dependências; `/health/ready` verifica MySQL e SNS a cada chamada, criando adapters próprios para a verificação. Nenhuma das rotas retorna corpo.

| Método | Caminho | Finalidade | Sucesso | Falha |
|---|---|---|---|---|
| `GET` | `/health/live` | Verifica se o processo está ativo | `200` (sem corpo) | — |
| `GET` | `/health/ready` | Verifica MySQL e SNS | `200` (sem corpo) | `503` (sem corpo) |

## Cenários de Usuário e Testes *(obrigatório)*

### História de Usuário 1 - Responder liveness (Prioridade: P1)

Um orquestrador precisa saber se o processo está no ar antes de rotear tráfego.

**Por que esta prioridade**: é a sonda mais barata e não depende de nada além do servidor HTTP.

**Teste independente**: subir o servidor e chamar `GET /health/live`; verificar `200` sem corpo.

**Cenários de aceitação**:

1. **Dado** o servidor no ar, **Quando** `GET /health/live` é chamado, **Então** a resposta é `200` com `c.NoContent(http.StatusOK)`.
2. **Dado** MySQL ou SNS indisponíveis, **Quando** `GET /health/live` é chamado, **Então** a resposta continua `200` — nenhuma dependência é consultada.

---

### História de Usuário 2 - Verificar dependências na prontidão (Prioridade: P1)

Um orquestrador precisa saber se o serviço pode atender requisições de domínio.

**Por que esta prioridade**: `POST /v1/ledger/events` depende de MySQL e SNS; a prontidão precisa refletir isso.

**Teste independente**: com dependências no ar, chamar `GET /health/ready` e verificar `200`; derrubar o MySQL e verificar `503`.

**Cenários de aceitação**:

1. **Dado** MySQL e SNS disponíveis, **Quando** `GET /health/ready` é chamado, **Então** a resposta é `200` sem corpo.
2. **Dado** `db.PingContext` bem-sucedido e `sns.GetTopicAttributes` bem-sucedido, **Quando** a verificação termina, **Então** o handler retorna `200`.
3. **Dado** MySQL indisponível, **Quando** `Repository.Health()` falha, **Então** a resposta é `503` e o SNS não é consultado.
4. **Dado** SNS indisponível, **Quando** `Topic.Health()` falha, **Então** a resposta é `503`.
5. **Dado** qualquer chamada, **Quando** a verificação começa, **Então** é criado um contexto com timeout de `500ms` e novos adapters de repositório e publicador, fechados ao final.

---

### História de Usuário 3 - Falhar fechado sem corpo (Prioridade: P2)

Consumidores das sondas dependem de um contrato simples e previsível.

**Por que esta prioridade**: qualquer corpo ou redirect quebraria sondas genéricas de Kubernetes/load balancer.

**Teste independente**: simular falha de dependência e conferir status `503` com corpo vazio.

**Cenários de aceitação**:

1. **Dado** uma dependência indisponível, **Quando** a prontidão falha, **Então** o status é `503` e o corpo está vazio.
2. **Dado** uma requisição sem autenticação, **Quando** qualquer sonda é chamada, **Então** ela é atendida normalmente — não há autenticação no serviço.

---

### Casos Limite

- `Readiness` usa `context.WithTimeout(..., 500*time.Millisecond)`; o tempo limite não é configurável.
- O repositório usado na prontidão é criado por `mysql.NewRepository(ctx)`; se a conexão falhar, o construtor devolve um repositório com `db == nil` (`OBS-02`).
- `message.New(ctx)` também é criado por chamada, sem reaproveitamento do publicador do fluxo de eventos.
- As rotas são registradas em `server.routes()` junto dos interceptadores de tracer, logger e OTel.

## Requisitos *(obrigatório)*

### Requisitos Funcionais

- **FR-001**: O sistema DEVE expor `GET /health/live` respondendo `200` sem corpo e sem consultar dependências. *(FR-12)*
- **FR-002**: O sistema DEVE expor `GET /health/ready` respondendo `200` quando MySQL e SNS estiverem disponíveis. *(FR-12)*
- **FR-003**: O sistema DEVE responder `503` sem corpo quando `Repository.Health()` ou `Topic.Health()` falhar. *(FR-12)*
- **FR-004**: O sistema DEVE verificar MySQL com `PingContext` e SNS com `GetTopicAttributes`. *(IR-04)*
- **FR-005**: O sistema DEVE limitar a verificação de prontidão a um contexto de `500ms`.

### Entidades Principais

- **Server**: `internal/infra/rest/server.Server`, com `routes()`, `Liveness` e `Readiness`.
- **Repository.Health()**: ping no MySQL obtido de `config.New().GetMySQLConnectionString()`.
- **Topic.Health()**: delega ao `Health()` do publisher SNS do core-sdk.

## Critérios de Sucesso *(obrigatório)*

### Resultados Mensuráveis

- **SC-001**: `GET /health/live` responde `200` com corpo vazio em 100% das chamadas com o servidor no ar.
- **SC-002**: `GET /health/ready` responde `200` somente quando ambas as verificações passam.
- **SC-003**: 100% das falhas de dependência resultam em `503` com corpo vazio.
- **SC-004**: Nenhuma sonda consulta dependências além de MySQL e SNS.

## Suposições

- Kubernetes/load balancer interpreta `200` como pronto e `503` como não pronto.
- As dependências locais seguem o descrito em `platform.md` (MySQL `3306`, LocalStack `4566`).
- Não há testes automatizados das sondas na suíte padrão; o comportamento foi lido diretamente do código.
