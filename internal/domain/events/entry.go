package events

import (
	"github.com/shopspring/decimal"
)

type Entry struct {
	EntryTypeID   int64
	Amount        decimal.Decimal
	DebitAccount  string
	CreditAccount string
	Description   string
}
