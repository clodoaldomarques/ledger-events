package mysql

import (
	"time"

	"github.com/clodoaldomarques/ledger-events/internal/domain/events"
	"github.com/shopspring/decimal"
)

type EventTable struct {
	EventID        string       `json:"event_id"`
	CorrelationID  string       `json:"correlation_id"`
	OrgID          string       `json:"org_id"`
	ProcessingCode string       `json:"processing_code"`
	ProgramID      int64        `json:"program_id"`
	AccountID      int64        `json:"account_id"`
	Description    string       `json:"description"`
	Producer       string       `json:"producer"`
	CreatedAt      time.Time    `json:"created_at"`
	Entries        []EntryTable `json:"entries"`
}

type EntryTable struct {
	EventID       string          `json:"event_id"`
	EntryTypeID   int64           `json:"event_type_id"`
	Amount        decimal.Decimal `json:"amount"`
	DebitAccount  string          `json:"debit_account,omitempty"`
	CreditAccount string          `json:"credit_account,omitempty"`
	Description   string          `json:"description"`
}

func buildEventTable(cid string, e events.Event) (EventTable, error) {
	evt := EventTable{
		EventID:        e.EventID,
		CorrelationID:  cid,
		OrgID:          e.OrgID,
		ProgramID:      e.ProgramID,
		ProcessingCode: e.ProcessingCode,
		AccountID:      e.AccountID,
		Description:    e.Description,
		Producer:       e.Producer,
		CreatedAt:      e.CreatedAt,
		Entries:        make([]EntryTable, 0, len(e.Entries)),
	}

	for _, en := range e.Entries {
		ent, err := buildEntryTable(e.EventID, en)
		if err != nil {
			return EventTable{}, err
		}
		evt.Entries = append(evt.Entries, ent)
	}

	return evt, nil
}

func buildEntryTable(eventID string, e events.Entry) (EntryTable, error) {
	ent := EntryTable{
		EventID:       eventID,
		EntryTypeID:   e.EntryTypeID,
		Amount:        e.Amount,
		DebitAccount:  e.DebitAccount,
		CreditAccount: e.CreditAccount,
		Description:   e.Description,
	}

	return ent, nil
}
