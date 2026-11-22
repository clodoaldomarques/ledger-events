# Especificações do Ledger Events

Specs geradas por **engenharia reversa** do código em `main` (baseline `396db3e`, 2026-11-21; diagramas em `abb0e85`). Elas descrevem o comportamento observado no baseline; a spec 005 é prescritiva e **já foi implementada**, migrando o código para a topologia hexagonal.

> Commit de referência: `396db3e` (`docs: atualiza README e adiciona AGENTS.md`).

## Estrutura

| Arquivo | Conteúdo |
|---|---|
| `constitution.md` | Princípios e restrições extraídos da forma como o projeto é implementado |
| `platform.md` | Contexto técnico: stack, variáveis de ambiente, MySQL, SNS, infra local, deploy e testes |
| `001-criacao-de-eventos/` | Recebimento, busca da config, cálculo, persistência e publicação do evento |
| `002-api-de-health/` | Liveness e readiness (MySQL + SNS) |
| `003-configuracao-e-infraestrutura/` | Env, Docker Compose, Makefile, Terraform, Kubernetes, benchmarks e perfis |
| `004-observabilidade-e-shutdown/` | Correlação, spans, logs, métricas e desligamento gracioso |
| `005-hexagonal-adapters/` | **Prescritiva**: migração para portas e adapters; IDs canônicos da família ledger/balances |
| `openapi.yaml` | Contrato HTTP legível por máquina, espelho de `001-criacao-de-eventos` e `002-api-de-health` |

As specs `001`-`004` são espelhos do baseline. Seus caminhos de código foram atualizados para o layout da spec 005 (`internal/core/ledger/**` e `internal/adapters/**`), mas os desvios `OBS-01`..`OBS-11` permanecem congelados. A `005` usa os mesmos identificadores `US1`-`US5`, `FR-001`..`FR-014` e `SC-001`..`SC-007` dos demais repos da família para permitir rastreabilidade cruzada.

Cada funcionalidade segue o formato Spec Kit:

- `spec.md` — cenários de usuário, requisitos funcionais (FR), entidades e critérios de sucesso (SC).
- `plan.md` — decisões técnicas já presentes no código, estrutura de arquivos e complexidade conhecida.
- `tasks.md` — sequência de implementação/reconstrução.

## Convenção das tarefas

- `[x]` = comportamento **já implementado** e verificado no commit de referência.
- `[ ]` = lacuna ou melhoria identificada na engenharia reversa (bug, teste ausente, comportamento divergente).

Os identificadores `US`, `FR`, `SC` e `T` seguem a nomenclatura do Spec Kit. Os IDs antigos da engenharia reversa (`FR-xx`, `IR-xx`, `BR-xx`, `NFR-xx`, `OBS-xx`) foram preservados entre parênteses em cada requisito e continuam válidos para rastreabilidade. `AS-IS` identifica comportamento verificado no código; `INFERIDO`, intenção deduzida dele.

## De onde veio cada arquivo

| Spec anterior | Destino no formato Spec Kit |
|---|---|
| `00-contexto.md` | `README.md`, `constitution.md`, `001-criacao-de-eventos/spec.md` |
| `01-requisitos.md` | specs `001`-`004` (IDs originais anotados em cada FR) |
| `02-arquitetura.md` | `001`/`002`/`004` plans e `platform.md` |
| `03-api.md` | `001-criacao-de-eventos/plan.md`, `002-api-de-health/spec.md`, `openapi.yaml` |
| `04-dominio.md` | `001-criacao-de-eventos/spec.md` e `plan.md` |
| `05-dados.md` | `platform.md` e `001-criacao-de-eventos/plan.md` |
| `06-configuracao-e-operacao.md` | `platform.md`, `003`/`004` specs e plans |

## Convenção arquitetural (spec 005)

- `internal/core/ledger/domain`: regras e tipos de domínio; sem dependência de infraestrutura.
- `internal/core/ledger/application`: caso de uso `Service`, testes e benchmarks.
- `internal/core/ledger/ports`: `EventUseCase` (entrada) e `EventRepository`, `EventConfigProvider`, `EventPublisher`, `HealthChecker` (saída); `mock.go` é gerado a partir de `outbound.go`.
- `internal/adapters/inbound/http`: Echo, handlers, DTOs, headers e mapeador de erros.
- `internal/adapters/outbound`: MySQL, Ledger Config e SNS.
- `cmd/main.go`: única raiz de composição; adapters são construídos uma vez e fechados no shutdown.
- `make arch-check` falha se o core importar adapters, AWS SDK, Echo, `config`, MySQL driver ou gobreaker.

## Uso para estudo

1. Leia `constitution.md` e `platform.md`.
2. Escolha uma funcionalidade e leia `spec.md` antes de olhar o código.
3. Compare `plan.md` com os arquivos citados.
4. Use `tasks.md` para identificar o que foi implementado e o que ficou como exercício.
5. Para mudanças de arquitetura, leia a spec canônica `005-hexagonal-adapters/` antes de alterar o layout de pacotes.

As specs são um espelho do código: se o código mudar, atualize também estes arquivos e o `plan.md` correspondente na mesma alteração.
