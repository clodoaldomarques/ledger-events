# Especificação da Funcionalidade: Adoção da Arquitetura Hexagonal — `ledger-events`

**Branch da funcionalidade**: `005-hexagonal-adapters`

**Criado**: 2026-11-22

**Repo**: `github.com/clodoaldomarques/ledger-events` (participante da família ledger/balances)

**Status**: Proposta — especificação **prescritiva** (estado desejado), ainda não implementada

**Entrada**: `docs/specs/constitution.md`, `docs/specs/platform.md`, specs `001`-`004`, `AGENTS.md` e engenharia reversa deste repo.

> **Escopo**: esta é a spec 005 aplicada a `ledger-events`. Os IDs `US1`-`US5`, `FR-001`..`FR-014` e `SC-001`..`SC-007` são mantidos idênticos nos repos da família para rastreabilidade; o conteúdo abaixo descreve **somente este repo**. Nenhuma mudança de comportamento externo está autorizada.

> **Diferença em relação às specs 001-004**: aquelas são espelhos descritivos do código em `396db3e`. Esta é uma spec **prescritiva**: descreve o estado alvo e será considerada concluída quando o código e os documentos de apoio forem ajustados aqui.

## Contexto e Motivação

`ledger-events` já respeita a direção `domain ← application ← infra`, mas as portas vivem em `internal/application/events/interfaces.go` com nomes genéricos (`Repository`, `ConfigProvider`, `Topic`). O handler `CreateEvent` constrói `configApi.New(ctx)`, `mysql.NewRepository(ctx)` e `message.New(ctx)` a cada requisição e os fecha com `defer`; `Readiness` cria outros dois adapters por chamada. A composição está espalhada entre handler, `server.New()` e `cmd/main.go`. As interfaces de domínio `Config`/`Script` são implementadas por DTOs que hoje moram em `internal/infra/ledger/config/response.go`.

Isso dificulta: (a) reusar o caso de uso em um segundo adaptador de entrada; (b) trocar MySQL/SNS/Ledger Config sem tocar nas regras; (c) manter a mesma convenção dos demais repos ledger/balances; (d) verificar a regra de dependência automaticamente. Há ainda uma dependência legada (`github.com/golang/mock`) usada apenas no teste do cliente de config e erros de domínio nunca chamados no fluxo.

A migração é estrutural: mover pacotes para `internal/core/ledger/{domain,application,ports}` e `internal/adapters/{inbound,outbound}`, nomear portas por intenção, concentrar a composição em `cmd/main.go` e remover o código morto. Regras de negócio, rotas HTTP, contrato SNS, DDL MySQL, spans e os desvios `OBS-01`..`OBS-11` permanecem congelados — cada correção deve virar spec própria.

## Cenários de Usuário e Testes *(obrigatório)*

### US1 — Trocar adaptador de saída sem alterar casos de uso (Prioridade: P1)

Como mantenedor de `ledger-events`, quero substituir MySQL, SNS ou o cliente HTTP do Ledger Config por outro adaptador (ou por um fake) sem modificar `services.go`.

**Por que esta prioridade**: é a garantia central da arquitetura hexagonal e o que torna a suíte determinística sem MySQL, LocalStack ou MockServer.

**Teste independente**: executar a mesma suíte de `services_test.go` com fakes in-memory de `EventRepository`, `EventConfigProvider` e `EventPublisher`; ambos devem satisfazer os mesmos contratos.

**Cenários de aceitação**:

1. **Dado** `internal/core/ledger/**`, **Quando** inspecionar seus imports, **Então** não há import de `internal/adapters`, AWS SDK, Echo, `config`, MySQL driver nem gobreaker.
2. **Dado** um fake que implementa as portas de saída, **Quando** injetado no `Service`, **Então** `CreateEvent` se comporta como hoje: config → cálculo → validação → persistência → publicação.
3. **Dado** o adaptador MySQL real, **Quando** compilado, **Então** satisfaz integralmente `EventRepository`.
4. **Dado** um método novo em uma porta, **Quando** algum adapter não o implementar, **Então** o compilador falha — sem fallback dinâmico.

---

### US2 — Adicionar adaptador de entrada sem duplicar regra (Prioridade: P1)

Como mantenedor, quero que o handler HTTP e um futuro consumer/CLI chamem o mesmo caso de uso pela porta de entrada `EventUseCase`.

**Por que esta prioridade**: hoje o handler depende de `*events.Service` e constrói seus próprios adapters; um segundo adaptador precisaria conhecer a implementação concreta e poderia replicar validação.

**Teste independente**: um adaptador de entrada alternativo invoca `EventUseCase` e verifica que o evento foi gravado e publicado, usando o mesmo `Service` do HTTP.

**Cenários de aceitação**:

1. **Dado** o handler HTTP e um segundo adaptador, **Quando** ambos chamam o caso de uso, **Então** utilizam a mesma implementação e as mesmas regras.
2. **Dado** um DTO de qualquer adaptador de entrada, **Quando** convertido para entidade, **Então** a validação continua em `Event.Validate()` e a geração de `event_id`/`created_at` permanece no mapeamento.
3. **Dado** um teste de handler, **Quando** ele injeta um fake de `EventUseCase`, **Então** não são necessários MySQL, SNS ou MockServer.

---

### US3 — Compor dependências uma única vez na raiz (Prioridade: P1)

Como operador/mantenedor, quero que os adapters sejam construídos no bootstrap e injetados, não recriados a cada requisição.

**Por que esta prioridade**: o handler chama três construtores por request e `Readiness` mais dois; isso espalha composição e dificulta controlar ciclo de vida e shutdown.

**Teste independente**: subir o servidor com adapters fake e verificar, por contador, uma instância de cada porta por processo e fechamento explícito no graceful shutdown.

**Cenários de aceitação**:

1. **Dado** `cmd/main.go`, **Quando** a aplicação inicia, **Então** cria `config.New()`, cliente MySQL, cliente HTTP instrumentado, `sns.NewPublisher`, adapters, `Service`, handlers e servidor, injetando as dependências.
2. **Dado** qualquer handler, **Quando** uma requisição chega, **Então** ele não chama `configApi.New`, `mysql.NewRepository` nem `message.New`.
3. **Dado** `/health/ready`, **Quando** chamado, **Então** consulta uma porta `HealthChecker` injetada em vez de instanciar adapters.
4. **Dado** o sinal de shutdown, **Quando** o graceful shutdown termina, **Então** os adapters com `Close()` são fechados pela raiz, preservando o comportamento atual (`Repository.Close()` vazio e `Topic.Close()` que apenas loga).

---

### US4 — Preservar contrato externo e levantar guarda arquitetural (Prioridade: P2)

Como revisor, quero garantir que a migração não altere API, eventos, DDL, spans ou modelo de dados, e que a regra de dependência continue válida.

**Por que esta prioridade**: a mudança é estrutural; qualquer diferença observável seria um incidente disfarçado de refatoração.

**Teste independente**: `go test ./...`, `go build ./...`, `go vet ./...`, `make arch-check` e comparação das respostas de `POST /v1/ledger/events` e `/health/*` com [`../openapi.yaml`](../openapi.yaml).

**Cenários de aceitação**:

1. **Dadas** as rotas atuais, **Quando** a migração termina, **Então** paths, status codes (`201`/`400`/`200`/`503`) e corpo `{"message": "..."}` permanecem idênticos.
2. **Dado** o contrato SNS, **Quando** comparado ao estado atual, **Então** campos e comportamentos conhecidos — inclusive as tags `entrytype_id`, o `make`+`append` em `ToEntryMessage` (`OBS-01`) e o `event_id` próprio do envelope — ficam congelados.
3. **Dado** o DDL, **Quando** comparado a `scripts/sql/init.sql`, **Então** nenhuma coluna, precisão ou chave muda.
4. **Dado** um import de `core` para `adapters`, **Quando** `make arch-check` roda, **Então** o comando falha com arquivo e linha infratores.
5. **Dado** o `git diff`, **Quando** revisado, **Então** `internal/core/ledger/domain` contém apenas movimento de pacote/imports, sem mudança de regra.

---

### US5 — Remover código morto e dependências órfãs (Prioridade: P2)

Como mantenedor, quero eliminar o que nunca entra no fluxo de request, reduzindo superfície de compilação e documentação.

**Por que esta prioridade**: `github.com/golang/mock` sobrevive apenas no teste do cliente de config, e `ErrDuplicatedEntry`/`ErrDuplicatedAccount` nunca são usados; ambos sobrevivem por inércia.

**Teste independente**: `go mod tidy` e busca por `golang/mock`, `ErrDuplicatedEntry` e `ErrDuplicatedAccount` neste repo não retornam ocorrências.

**Cenários de aceitação**:

1. **Dado** `internal/infra/ledger/config/ledger_config_test.go`, **Quando** a tarefa é executada, **Então** ele passa a usar `go.uber.org/mock/gomock`.
2. **Dado** `go.mod`, **Quando** `go mod tidy` roda, **Então** `github.com/golang/mock` desaparece das dependências diretas.
3. **Dado** `internal/domain/events/errors.go`, **Quando** a limpeza termina, **Então** `ErrDuplicatedEntry` e `ErrDuplicatedAccount` são removidos; `ErrScriptNotFound` e `ErrEntryNotFound` permanecem.
4. **Dada** alguma funcionalidade do fluxo de request, **Quando** comparada ao estado anterior, **Então** não houve alteração — o código removido estava inalcançável.

---

### Casos Limite

- Adapters passam a ser compartilhados por todas as requisições: confirmar que `sns.Publisher`, `sql.DB` e o `httpclient` do core-sdk são seguros para uso concorrente antes da Fase 3.
- `config.New()` continua `sync.Once`; mudar esse comportamento está fora do escopo. `APP_PORT` continua com default 5001.
- `x-cid`/`x-tenant` e a geração automática de correlation id permanecem em `adapters/inbound/http/shared`.
- `tracer`/`logger` hoje são chamados dentro do caso de uso e dos adapters. Nesta spec **permanecem**; movê-los para decorators é fase opcional futura.
- Os `Close()` atuais (`Repository.Close()` vazio, `Topic.Close()` que loga) devem ser preservados ou removidos explicitamente na Fase 3, sem alterar respostas.
- Os desvios `OBS-01`..`OBS-11` catalogados nas specs 001-004 **não** são corrigidos aqui.
- O mapeador de erros continua respondendo `400` para qualquer falha (`OBS-06`); a separação `4xx`/`5xx` exige spec própria.
- `make test` abre o relatório HTML e não deve ser usado em checagens não interativas; usar `go test ./...`.
- A collection em `docs/collections/` está desatualizada (`OBS-08`); usar `../openapi.yaml` como referência de contrato nesta spec.

## Requisitos *(obrigatório)*

### Requisitos Funcionais

- **FR-001**: O core (`internal/core/ledger/**`) NÃO DEVE importar `internal/adapters`, AWS SDK, Echo, `config`, MySQL driver, gobreaker nem outros pacotes de infraestrutura.
- **FR-002**: As portas DEVEM ser declaradas em `internal/core/ledger/ports` e implementadas pelos adapters; a dependência aponta sempre do adapter para o core.
- **FR-003**: O caso de uso DEVE ser exposto por `EventUseCase` no core, consumido pelos adaptadores de entrada.
- **FR-004**: A raiz de composição (`cmd/main.go`) DEVE instanciar cada adapter uma única vez e injetá-lo no `Service` e nos handlers.
- **FR-005**: Os handlers NÃO DEVEM construir adapters nem casos de uso; recebem as portas por construtor.
- **FR-006**: Health checks DEVEM ser expostos como porta `HealthChecker` injetada e consumida pelo adaptador HTTP em `/health/ready`.
- **FR-007**: O mapeamento de erros do domínio para status HTTP (`400`) DEVE ficar em `adapters/inbound/http/errors.go`.
- **FR-008**: As portas DEVEM ter nomes que expressem intenção: `EventUseCase`, `EventRepository`, `EventConfigProvider`, `EventPublisher`, `HealthChecker` — em vez de `Repository`/`ConfigProvider`/`Topic`.
- **FR-009**: A dependência legada `github.com/golang/mock` DEVE ser removida após migrar o teste para `go.uber.org/mock`; `ErrDuplicatedEntry` e `ErrDuplicatedAccount` DEVEM ser removidos.
- **FR-010**: O comportamento observável NÃO PODE mudar: rotas, status, JSON, expressões, chaves do MySQL, tópico SNS, envelope de evento, spans, timeouts de 500ms/25s e endpoint do LocalStack.
- **FR-011**: DEVE existir `make arch-check` neste repo, executado antes do merge.
- **FR-012**: A constituição, `platform.md`, `docs/specs/README.md` e os planos das specs `001`-`004` afetados DEVEM ser atualizados na mesma mudança de código.
- **FR-013**: A remoção de código morto NÃO DEVE alterar funcionalidade observável; nenhum dos itens removidos faz parte do fluxo de request.
- **FR-014**: Este repo DEVE manter um único `spec.md`, `plan.md` e `tasks.md` coerentes com o layout alvo, mesmo que a implementação ocorra em PRs separados.

### Entidades Principais

- **Porta de entrada (driving port)**: `EventUseCase` no core, com o método público `CreateEvent` do `Service` atual.
- **Portas de saída (driven ports)**: `EventRepository`, `EventConfigProvider`, `EventPublisher` e `HealthChecker` no core.
- **Adapter de entrada**: handlers HTTP + `server.New()` + `shared` (DTOs, `x-cid`, mapeador de erros).
- **Adapter de saída**: `adapters/outbound/{mysql,config,sns}`.
- **Composition root**: `cmd/main.go`; único lugar autorizado a conhecer adapters concretos e configuração de ambiente.
- **Guarda de arquitetura**: `scripts/check-architecture.sh`, exposta por `make arch-check`.

## Critérios de Sucesso *(obrigatório)*

### Resultados Mensuráveis

- **SC-001**: `go test ./...`, `go build ./...` e `go vet ./...` passam neste repo antes e depois da migração.
- **SC-002**: Zero imports de `internal/adapters` a partir de `internal/core` neste repo.
- **SC-003**: Zero chamadas a `configApi.New`, `mysql.NewRepository` ou `message.New` fora de `cmd/**` e dos próprios adapters.
- **SC-004**: Diff de `internal/core/ledger/domain` restrito a movimento de pacote/imports — nenhuma regra alterada.
- **SC-005**: Nenhuma diferença nas respostas de `POST /v1/ledger/events` e `/health/*` comparadas a `../openapi.yaml`.
- **SC-006**: `go mod tidy` sem `github.com/golang/mock` e sem ocorrências de `ErrDuplicatedEntry`/`ErrDuplicatedAccount` neste repo.
- **SC-007**: `make arch-check` falha quando um import proibido é introduzido artificialmente e passa no estado final.

## Suposições

- Os clientes AWS do `core-sdk`, o `sql.DB` do MySQL driver e o `httpclient` do core-sdk são seguros para reutilização concorrente; confirmar na Fase 0 antes de compartilhar instâncias.
- O contrato externo atual é o correto, inclusive os desvios registrados nas specs 001-004; correções viram specs separadas.
- Este repo é um módulo Go independente (`go.mod` próprio).
- Toda a documentação necessária para este repo está neste diretório e nos arquivos `README.md`/`platform.md`/`constitution.md` da raiz de specs.
- Não há CI; a guarda arquitetural é executada por comando local (`make arch-check`) e pelo revisor.
