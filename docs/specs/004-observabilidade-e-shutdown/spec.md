# Especificação da Funcionalidade: Observabilidade e Desligamento Gracioso

**Branch da funcionalidade**: `004-observabilidade-e-shutdown`

**Criado**: 2026-11-22

**Status**: Rascunho (derivado do código em `396db3e`)

**Entrada**: engenharia reversa de `cmd/main.go`, `internal/adapters/inbound/http/server/server.go`, `internal/core/ledger/application/services.go`, `internal/adapters/outbound/{mysql,config,sns}` e `scripts/docker/otel-collector.yaml` (caminhos originais do baseline `396db3e` atualizados para o layout da spec 005)

> **Escopo**: espelho `AS-IS` de correlação, spans, logs, métricas e ciclo de vida do processo. Variáveis OTEL e ports do coletor estão em [`../platform.md`](../platform.md).

## Visão Geral

O `main` inicia o OpenTelemetry antes do servidor e o encerra após o graceful shutdown. Todo request passa por interceptadores de tracer, logger e instrumentação OTel do `core-sdk`. O correlation id `x-cid` é propagado como atributo `cid` nos spans do caso de uso e dos adapters.

## Cenários de Usuário e Testes *(obrigatório)*

### História de Usuário 1 - Correlacionar requisições (Prioridade: P1)

Um operador precisa ligar uma requisição HTTP aos spans, logs e registros de banco correspondentes.

**Por que esta prioridade**: sem correlation id, o rastreamento distribuído perde utilidade no caminho síncrono e assíncrono.

**Teste independente**: enviar `x-cid` conhecido e verificar o atributo `cid` nos spans e a coluna `correlation_id` no MySQL.

**Cenários de aceitação**:

1. **Dado** `x-cid` no header, **Quando** o handler lê os headers, **Então** o valor é passado como `cid` para `Service.CreateEvent`.
2. **Dado** `x-cid` ausente, **Quando** o handler lê os headers, **Então** `uuid.NewString()` é gerado e usado.
3. **Dado** um `cid`, **Quando** os spans são criados no caso de uso e nos adapters, **Então** todos carregam o atributo `cid`.
4. **Dado** um evento gravado, **Quando** `SaveEvent` é executado, **Então** o `cid` aparece em `event.correlation_id`.

---

### História de Usuário 2 - Emitir spans, logs e métricas (Prioridade: P1)

Um operador precisa observar o caminho de criação de eventos em traces, logs e métricas.

**Por que esta prioridade**: é a principal ferramenta de diagnóstico de um serviço que integra HTTP, MySQL, HTTP externo e SNS.

**Teste independente**: executar `make run` com o coletor local e verificar no Zipkin os quatro spans nomeados e no Prometheus as métricas do coletor.

**Cenários de aceitação**:

1. **Dado** o servidor iniciado, **Quando** uma requisição chega, **Então** os interceptadores `tracer.Interceptor()`, `logger.Interceptor()` e `otelecho.Interceptor()` são aplicados.
2. **Dado** o caso de uso, **Quando** `CreateEvent` é executado, **Então** é criado o span `Service::CreateEvent`.
3. **Dado** a consulta de config, **Quando** `FindConfigByLevel` é executado, **Então** é criado o span `LedgerConfigApi::FindConfigByLevel`.
4. **Dado** a persistência, **Quando** `SaveEvent` é executado, **Então** é criado o span `Repository::SaveEvent`.
5. **Dado** a publicação, **Quando** `Emit` é executado, **Então** é criado o span `Topic::Emit`.
6. **Dado** logger Zap do core-sdk, **Quando** eventos de lifecycle ocorrem, **Então** são registrados de forma estruturada; `Topic.Close(ctx)` registra `ending topic connection`.
7. **Dado** o coletor OTLP, **Quando** recebe os dados, **Então** envia traces ao Zipkin e expõe métricas ao Prometheus.

---

### História de Usuário 3 - Iniciar e encerrar a telemetria com o processo (Prioridade: P2)

Um operador precisa que a telemetria esteja disponível desde o startup e liberada no shutdown.

**Por que esta prioridade**: perder spans no final do processo tornaria o rastreamento incompleto justamente durante incidentes de desligamento.

**Teste independente**: iniciar o binário, enviar `SIGTERM` e conferir as mensagens de shutdown e a ordem HTTP → OTel.

**Cenários de aceitação**:

1. **Dado** o `main`, **Quando** a aplicação inicia, **Então** `otel.Start(ctx)` é chamado antes de `server.New()`.
2. **Dado** `SIGINT` ou `SIGTERM`, **Quando** o sinal é recebido, **Então** o `main` cria contexto com timeout de `25s`.
3. **Dado** o contexto de shutdown, **Quando** o processo termina, **Então** `Server.Shutdown` é chamado antes de `otel.Shutdown`.
4. **Dado** erro em qualquer etapa do shutdown, **Quando** ele ocorre, **Então** a mensagem é impressa no stdout e o processo termina com status `0`.
5. **Dado** uma requisição em andamento, **Quando** o shutdown começa, **Então** o servidor Echo tenta encerrá-la dentro dos `25s`.

---

### Casos Limite

- O timeout de shutdown é fixo em `25s`, não configurável.
- Erros de `Start` diferentes de `http.ErrServerClosed` chamam `logger.Fatal`.
- Os clients de request (`configApi`, repositório, topic) são fechados com `defer` no handler; `Topic.Close(ctx)` apenas loga.
- O `NewSpanFromContext` do core-sdk é chamado dentro do caso de uso e dos adapters; movê-lo para decorators está fora do escopo (avaliado na spec 005).
- Não há teste automatizado de telemetria na suíte padrão.

## Requisitos *(obrigatório)*

### Requisitos Funcionais

- **FR-001**: O sistema DEVE gerar correlation id quando `x-cid` estiver ausente. *(FR-02)*
- **FR-002**: O sistema DEVE propagar `cid` em spans, logs, banco e headers. *(NFR-01)*
- **FR-003**: O sistema DEVE criar os spans `Service::CreateEvent`, `LedgerConfigApi::FindConfigByLevel`, `Repository::SaveEvent` e `Topic::Emit`. *(NFR-01)*
- **FR-004**: O sistema DEVE registrar tracer, logger e instrumentação OTel como middlewares do Echo. *(NFR-02)*
- **FR-005**: O sistema DEVE chamar `otel.Start(ctx)` no startup e `otel.Shutdown(ctx)` no fim do `main`. *(FR-14)*
- **FR-006**: O sistema DEVE tratar `SIGINT`/`SIGTERM` com contexto de timeout de `25s`. *(FR-13)*
- **FR-007**: O sistema DEVE encerrar o HTTP antes da telemetria no graceful shutdown. *(FR-13)*

### Entidades Principais

- **Tracer do core-sdk**: `tracer.NewSpanFromContext(ctx, nome, atributos)`.
- **Interceptadores Echo**: `tracer.Interceptor()`, `logger.Interceptor()`, `otelecho.Interceptor()`.
- **Logger**: `core-sdk/pkg/zap/logger`, com `Fields` estruturados.
- **Main**: `cmd/main.go`, responsável por `otel.Start`, `server.Start`, sinais e shutdown.

## Critérios de Sucesso *(obrigatório)*

### Resultados Mensuráveis

- **SC-001**: 100% dos spans do caminho de criação de eventos contêm o atributo `cid`.
- **SC-002**: O correlation id informado ou gerado aparece em `event.correlation_id` em 100% dos eventos gravados.
- **SC-003**: Os quatro spans nomeados são visíveis no Zipkin após uma requisição bem-sucedida.
- **SC-004**: Após `SIGTERM`, o HTTP é encerrado antes do OTel e o processo termina com status `0`.
- **SC-005**: Métricas do coletor estão disponíveis no Prometheus na porta `9090`.

## Suposições

- O coletor OTLP, Zipkin e Prometheus estão disponíveis conforme `platform.md`.
- O `core-sdk` fornece interceptadores, tracer e logger compatíveis com a versão em `go.mod`.
- Não há SLO de latência declarado; os spans servem para diagnóstico.
- Mudanças na estrutura de spans exigem atualização de `spec.md`, `plan.md` e das specs 001/003 quando afetarem contrato ou configuração.
