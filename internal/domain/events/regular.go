package events

import (
	"github.com/clodoaldomarques/core-sdk/pkg/expression"
	"github.com/clodoaldomarques/ledger-events/internal/domain/configs"
	"github.com/shopspring/decimal"
)

func ProcessRegular(c configs.Config, e *Event, amounts, fees map[string]decimal.Decimal) error {
	e.Description = c.Description

	for _, en := range c.RetrieveEntryByProducer(configs.Regular) {
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
