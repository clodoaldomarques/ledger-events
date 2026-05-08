package ledger

import (
	"time"

	"github.com/clodoaldomarques/ledger-events/internal/domain/events"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type EventRequest struct {
	ProcessingCode string                     `json:"processing_code" validate:"required"`
	ProgramID      int64                      `json:"program_id" validate:"required"`
	AccountID      int64                      `json:"account_id" validate:"required"`
	Producer       string                     `json:"producer" validate:"required"`
	Amounts        map[string]decimal.Decimal `json:"amounts" validate:"required"`
	Fees           map[string]decimal.Decimal `json:"fees" validate:"required"`
}

func (e EventRequest) Validate() error {
	return nil
}

func (e EventRequest) ToEntity(orgID string) events.Event {
	evt := events.Event{
		EventID:        uuid.NewString(),
		ProcessingCode: e.ProcessingCode,
		OrgID:          orgID,
		ProgramID:      e.ProgramID,
		AccountID:      e.AccountID,
		Producer:       e.Producer,
		CreatedAt:      time.Now(),
	}
	return evt
}
