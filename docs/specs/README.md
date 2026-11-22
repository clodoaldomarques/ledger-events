# Especificações — Ledger Events

Pacote de engenharia reversa do comportamento **atual** do serviço, reconstruído a partir do código, configurações e infraestrutura versionada.

- **Baseline:** commit `396db3e` (21/11/2026).
- **Idioma:** português (pt-BR).
- **Fonte de verdade:** código executável. Ocorrendo conflito entre README, Postman e código, prevalece o código.

## Convenções

| Prefixo | Uso |
|---|---|
| `FR-xx` | Requisito funcional |
| `NFR-xx` | Requisito não funcional |
| `BR-xx` | Regra de negócio / domínio |
| `IR-xx` | Integração com sistema externo |
| `OBS-xx` | Comportamento observado que pode ser um desvio ou dívida técnica |

Cada requisito traz uma coluna de rastreabilidade apontando o arquivo ou função de origem.

Status usados nas specs:

- `AS-IS` — comportamento implementado e verificado diretamente no código.
- `INFERIDO` — intenção deduzida do código; pode divergir do desenho original.

## Índice

1. [Contexto e escopo](00-contexto.md)
2. [Requisitos](01-requisitos.md)
3. [Arquitetura e fluxos](02-arquitetura.md)
4. [Contrato HTTP](03-api.md) + [`openapi.yaml`](openapi.yaml)
5. [Modelo de domínio e regras de negócio](04-dominio.md)
6. [Modelo de dados](05-dados.md)
7. [Configuração, operação e dívidas conhecidas](06-configuracao-e-operacao.md)

## Uso em um fluxo SDD

1. **Especificar:** ajuste um requisito neste pacote antes de tocar no código.
2. **Planejar:** registre a mudança de comportamento esperada e os casos de teste que a cobrem.
3. **Implementar:** altere o código e mantenha a coluna de rastreabilidade atualizada.
4. **Verificar:** `go test ./...`, `go vet ./...` e, quando aplicável, `go generate ./internal/application/events`.
5. **Promover:** se um `OBS-xx` for corrigido, transforme-o em requisito com critério de aceitação e remova a observação.
