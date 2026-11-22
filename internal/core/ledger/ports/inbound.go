package ports

import (
	"context"

	"github.com/clodoaldomarques/ledger-events/internal/core/ledger/domain"
	"github.com/shopspring/decimal"
)

// EventUseCase é a porta de entrada do caso de uso de criação de eventos.
type EventUseCase interface {
	CreateEvent(ctx context.Context, cid string, e domain.Event, amounts, fees map[string]decimal.Decimal) (domain.Event, error)
}
