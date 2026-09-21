package events

import (
	"github.com/clodoaldomarques/core-sdk/pkg/expression"
	"github.com/shopspring/decimal"
)

func ProcessMigration(c Config, e *Event, amounts, fees map[string]decimal.Decimal) error {
	e.Description = c.RetrieveDescription()

	for _, en := range c.RetrieveEntryByProducer(Migration) {
		calculated, err := expression.Calculate(en.RetrieveExpression(), amounts, fees)
		if err != nil {
			return err
		}

		entry := Entry{
			EntryTypeID: en.RetrieveScriptID(),
			Amount:      calculated,
			Description: en.RetrieveDescription(),
		}

		entry.CreditAccount = en.RetrieveCreditAccount()
		entry.DebitAccount = en.RetrieveDebitAccount()

		e.Entries = append(e.Entries, entry)
	}

	return nil
}
