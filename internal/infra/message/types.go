package message

import (
	"time"

	"github.com/clodoaldomarques/ledger-events/internal/domain/events"
	"github.com/shopspring/decimal"
)

type Event struct {
	EventID        string    `json:"event_id"`
	OrgID          string    `json:"org_id"`
	ProcessingCode string    `json:"processing_code"`
	ProgramID      int64     `json:"program_id"`
	AccountID      int64     `json:"account_id"`
	Description    string    `json:"description"`
	Producer       string    `json:"producer"`
	CreatedAt      time.Time `json:"created_at"`
	Entries        []Entry   `json:"entries"`
}

type Entry struct {
	EntryTypeID   int64           `json:"entrytype_id"`
	Amount        decimal.Decimal `json:"amount"`
	DebitAccount  string          `json:"debit_account_id"`
	CreditAccount string          `json:"credit_account_id"`
	Description   string          `json:"description"`
}

func ToEventMessage(e events.Event) Event {
	return Event{
		EventID:        e.EventID,
		OrgID:          e.OrgID,
		ProcessingCode: e.ProcessingCode,
		ProgramID:      e.ProgramID,
		AccountID:      e.AccountID,
		Description:    e.Description,
		Producer:       e.Producer,
		CreatedAt:      e.CreatedAt,
		Entries:        ToEntryMessage(e.Entries),
	}
}

func ToEntryMessage(et []events.Entry) []Entry {
	em := make([]Entry, len(et))
	for _, e := range et {
		en := Entry{
			EntryTypeID:   e.EntryTypeID,
			Amount:        e.Amount,
			DebitAccount:  e.DebitAccount,
			CreditAccount: e.CreditAccount,
			Description:   e.Description,
		}
		em = append(em, en)
	}
	return em
}
