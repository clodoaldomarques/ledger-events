package ledger

import (
	"time"

	"github.com/clodoaldomarques/ledger-events/internal/domain/events"
	"github.com/shopspring/decimal"
)

type EventResponse struct {
	EventID        string          `json:"event_id"`
	OrgID          string          `json:"org_id"`
	ProcessingCode string          `json:"processing_code"`
	ProgramID      int64           `json:"program_id"`
	AccountID      int64           `json:"account_id"`
	Description    string          `json:"description"`
	Producer       string          `json:"producer"`
	CreatedAt      time.Time       `json:"created_at"`
	Entries        []EntryResponse `json:"entries"`
}

func buildEventResponse(e events.Event) EventResponse {
	er := EventResponse{
		EventID:        e.EventID,
		OrgID:          e.OrgID,
		ProcessingCode: e.ProcessingCode,
		ProgramID:      e.ProgramID,
		AccountID:      e.AccountID,
		Description:    e.Description,
		Producer:       e.Producer,
		CreatedAt:      e.CreatedAt,
	}

	for _, en := range e.Entries {
		er.Entries = append(er.Entries, buildEntryResponse(en))
	}

	return er
}

type EntryResponse struct {
	EntryTypeID   int64           `json:"entry_type_id"`
	Amount        decimal.Decimal `json:"amount"`
	DebitAccount  string          `json:"debit_account"`
	CreditAccount string          `json:"credit_account"`
	Description   string          `json:"description"`
}

func buildEntryResponse(e events.Entry) EntryResponse {
	er := EntryResponse{
		EntryTypeID:   e.EntryTypeID,
		Amount:        e.Amount,
		DebitAccount:  e.DebitAccount,
		CreditAccount: e.CreditAccount,
		Description:   e.Description,
	}

	return er
}
