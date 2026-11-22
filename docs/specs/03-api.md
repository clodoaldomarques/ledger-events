# 03 — Contrato HTTP

Base URL local: `http://localhost:5001`. Todas as requisições e respostas usam `application/json`, exceto os health checks, que não têm corpo.

A versão legível por máquina deste contrato está em [`openapi.yaml`](openapi.yaml).

## Endpoints

| Método | Caminho | Finalidade | Sucesso | Falha |
|---|---|---|---|---|
| `GET` | `/health/live` | Verifica se o processo está ativo | `200` (sem corpo) | — |
| `GET` | `/health/ready` | Verifica MySQL e SNS | `200` (sem corpo) | `503` (sem corpo) |
| `POST` | `/v1/ledger/events` | Cria e publica um evento contábil | `201` | `400` |

## `POST /v1/ledger/events`

### Headers

| Header | Obrigatório na prática | Descrição |
|---|---|---|
| `x-tenant` | Não validado, mas esperado | Identificador da organização; vira `org_id` e é reenviado ao Ledger Config. |
| `x-cid` | Não | Correlation id. Quando ausente, o servidor gera um UUID. |

`Content-Type: application/json` é necessário para o bind do corpo.

### Corpo da requisição

| Campo | Tipo | Descrição |
|---|---|---|
| `processing_code` | string | Código da config no Ledger Config. |
| `program_id` | integer (int64) | Complemento da chave de configuração. |
| `account_id` | integer (int64) | Conta associada ao evento. |
| `producer` | string | `regular` ou `migration`. |
| `amounts` | objeto de decimais | Valores base das expressões, acessíveis como `Amounts`. |
| `fees` | objeto de decimais | Taxas das expressões, acessíveis como `Fees`. |

Exemplo compatível com as expectativas do MockServer local (`/v1/ledger/config/101/401`):

```json
{
  "processing_code": "101",
  "program_id": 401,
  "account_id": 123,
  "producer": "regular",
  "amounts": { "amount": 150.00 },
  "fees": { "iof": 15.99 }
}
```

Valores decimais podem vir como número ou string; a resposta serializa decimais como string (ex.: `"165.99"`).

### Resposta `201`

```json
{
  "event_id": "9f1d...",
  "org_id": "TN-77add76c-e395-446b-b306-1a0f9cb99a31",
  "processing_code": "101",
  "program_id": 401,
  "account_id": 123,
  "description": "Compra a vista",
  "producer": "regular",
  "created_at": "2026-11-21T20:00:00Z",
  "entries": [
    {
      "entry_type_id": 101,
      "amount": "165.99",
      "debit_account": "",
      "credit_account": "",
      "description": "Compra a vista - Cartão"
    }
  ]
}
```

`event_id` e `created_at` são gerados pelo servidor. `description` vem da config. As contas vêm dos campos `debit_account.number` e `credit_account.number` do script; são strings vazias quando o script não as informa.

### Erros `400`

Formato único: `{"message": "<texto>"}`.

| Situação | Mensagem | Origem |
|---|---|---|
| Falha ao vincular o corpo | `Bad Request` | Echo / `echo.ErrBadRequest` |
| Config não encontrada (`404` no Ledger Config) | `ledger config not found` | `ErrResponse.Message` do cliente de config |
| Expressão inválida | `invalid expression: <expressão>` | core-sdk `expression.ErrInvalidExpression` |
| Erro ao avaliar expressão | `error on calculate: <detalhe>` | core-sdk `expression.ErrCalculateExpression` |
| Evento sem entradas | `entry not found to producer <producer>` | `domain/events.ErrEntryNotFound` |
| Config com status inesperado | `api error: status <código>` | `ledger_config.go` |
| JSON inválido vindo do Ledger Config | `unmarshal error: ...` / `unmarshal erro: ...` | `ledger_config.go` (observe a grafia inconsistente) |
| Erro de MySQL ou SNS | mensagem retornada pelo driver / core-sdk | adapters de infraestrutura |

Todas essas falhas, inclusive de dependências externas, usam `400` (ver `OBS-06`).

### Health checks

- `GET /health/live` — sempre `200`, desde que o servidor esteja no ar.
- `GET /health/ready` — `200` quando `db.PingContext` e `sns.GetTopicAttributes` passam; `503` quando qualquer um falha. Sem corpo.

## Observações

- Não há autenticação ou autorização no serviço.
- Não há endpoint de consulta de eventos.
- A collection Postman em `docs/collections/` usa `/v1/accounting/events` e campos antigos; não a use como contrato (ver `OBS-08`).
- Um script com `flow` diferente de `producer` é ignorado; se nenhum script casar, o evento não terá entradas e falha na validação de domínio.
