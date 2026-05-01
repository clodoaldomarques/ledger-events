package accounting

import (
	"time"

	"github.com/clodoaldomarques/ledger-events/internal/domain/events"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type EventRequest struct {
	EventTypeID     string                     `json:"event_type_id" validate:"required"`
	TransactionID   *int64                     `json:"transaction_id,omitempty"`
	AuthorizationID *int64                     `json:"authorization_id,omitempty"`
	ProgramID       int64                      `json:"program_id" validate:"required"`
	CustomerID      *int64                     `json:"customer_id,omitempty"`
	AccountID       int64                      `json:"account_id" validate:"required"`
	DaysDue         int64                      `json:"days_due"`
	EventDate       time.Time                  `json:"event_date" validate:"required"`
	TransactionType int64                      `json:"transaction_type"`
	ProcessID       int64                      `json:"process_id"`
	Producer        string                     `json:"producer" validate:"required"`
	Processing      *ProcessingRequest         `json:"processing,omitempty"`
	ClearingDate    *time.Time                 `json:"clearing_date,omitempty"`
	ChunkPart       int64                      `json:"chunk_part" validate:"required"`
	Amounts         map[string]decimal.Decimal `json:"amounts" validate:"required"`
	Fees            map[string]decimal.Decimal `json:"fees" validate:"required"`
}

func (e EventRequest) Validate() error {
	if e.Processing != nil {
		if err := e.Processing.Validate(); err != nil {
			return err
		}
	}
	return nil
}

func (e EventRequest) ToEntity(orgID string) events.Event {
	evt := events.Event{
		EventID:         uuid.NewString(),
		EventTypeID:     e.EventTypeID,
		TransactionID:   e.TransactionID,
		AuthorizationID: e.AuthorizationID,
		OrgID:           orgID,
		ProgramID:       e.ProgramID,
		CustomerID:      e.CustomerID,
		AccountID:       e.AccountID,
		DaysDue:         e.DaysDue,
		EventDate:       e.EventDate,
		TransactionType: e.TransactionType,
		ProcessID:       e.ProcessID,
		Producer:        e.Producer,
		ChunkPart:       e.ChunkPart,
		CreatedAt:       time.Now(),
	}

	if e.Processing != nil {
		evt.Processing = e.Processing.ToEntity()
	}

	return evt
}

type ProcessingRequest struct {
	Code        string `json:"code" validate:"required"`
	Description string `json:"description" validate:"required"`
}

func (p ProcessingRequest) ToEntity() *events.Processing {
	return &events.Processing{
		Code:        p.Code,
		Description: p.Description,
	}
}

func (p ProcessingRequest) Validate() error {
	return nil
}
