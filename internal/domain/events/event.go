package events

import (
	"errors"
	"time"

	"github.com/clodoaldomarques/ledger-events/internal/domain/expression"
	"github.com/clodoaldomarques/ledger-events/internal/domain/scripts"
	"github.com/shopspring/decimal"
)

type Event struct {
	EventID         string
	OrgID           string
	EventTypeID     string
	TransactionID   *int64
	AuthorizationID *int64
	ProgramID       int64
	CustomerID      *int64
	AccountID       int64
	DaysDue         int64
	EventDate       time.Time
	Description     string
	TransactionType int64
	ProcessID       int64
	Producer        string
	Processing      *Processing
	ClearingDate    *time.Time
	ChunkPart       int64
	CreatedAt       time.Time
	Entries         []Entry
}

func (e Event) Validate() error {
	if e.TransactionID == nil && e.AuthorizationID == nil {
		return errors.New("transaction_id or authorization_id is required")
	}
	return nil
}

func (e *Event) Process(scr scripts.Script, amount, fee map[string]decimal.Decimal) error {
	e.Description = scr.Description

	for _, en := range scr.RetrieveEntryByProducer(e.Producer) {
		calculated, err := expression.Calculate(en.RetrieveExpr(), amount, fee)
		if err != nil {
			return err
		}

		entry := Entry{
			EntryTypeID: en.EntryTypeID,
			Amount:      calculated,
			Description: en.Description,
		}

		if scr.Company != nil {
			entry.Company = Company{scr.Company.Code, scr.Company.Type}
		}

		if en.CostCenter != nil {
			entry.CostCenter = &CostCenter{
				DebitOrg:   en.CostCenter.DebitOrg,
				DebitCost:  en.CostCenter.DebitCost,
				CreditOrg:  en.CostCenter.CreditOrg,
				CreditCost: en.CostCenter.CreditCost,
			}
		}

		if en.CreditAccount != nil {
			entry.CreditAccount = en.CreditAccount.Number
		}

		if en.DebitAccount != nil {
			entry.DebitAccount = en.DebitAccount.Number
		}

		e.Entries = append(e.Entries, entry)
	}

	return nil
}
