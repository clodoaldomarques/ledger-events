package events

import (
	"time"

	"github.com/clodoaldomarques/ledger-events/internal/domain/configs"
	"github.com/clodoaldomarques/ledger-events/internal/domain/expression"
	"github.com/shopspring/decimal"
)

type Event struct {
	EventID        string
	OrgID          string
	ProcessingCode string
	ProgramID      int64
	AccountID      int64
	Description    string
	Producer       string
	CreatedAt      time.Time
	Entries        []Entry
}

func (e Event) Validate() error {
	return nil
}

func (e *Event) Process(scr configs.Config, amounts, fees map[string]decimal.Decimal) error {
	e.Description = scr.Description

	for _, en := range scr.RetrieveEntryByProducer(e.Producer) {
		calculated, err := expression.Calculate(en.Expression, amounts, fees)
		if err != nil {
			return err
		}

		entry := Entry{
			EntryTypeID: en.ScriptID,
			Amount:      calculated,
			Description: en.Description,
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
