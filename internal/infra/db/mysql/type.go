package mysql

import (
	"time"

	"github.com/clodoaldomarques/ledger-events/internal/domain/events"
	"github.com/shopspring/decimal"
)

type Event struct {
	EventID               string     `json:"event_id"`
	CorrelationID         string     `json:"correlation_id"`
	OrgID                 string     `json:"org_id"`
	EventTypeID           string     `json:"event_type_id"`
	TransactionID         *int64     `json:"transaction_id"`
	AuthorizationID       *int64     `json:"authorization_id"`
	ProgramID             int64      `json:"program_id"`
	CustomerID            *int64     `json:"customer_id"`
	AccountID             int64      `json:"account_id"`
	DaysDue               int64      `json:"days_due"`
	EventDate             time.Time  `json:"event_date"`
	Description           string     `json:"description"`
	TransactionType       int64      `json:"transaction_type"`
	ProcessID             int64      `json:"process_id"`
	Producer              string     `json:"producer"`
	ProcessingCode        string     `json:"processing_code"`
	ProcessingDescription string     `json:"processing_description"`
	ClearingDate          *time.Time `json:"clearing_date"`
	ChunkPart             int64      `json:"chunk_part"`
	CreatedAt             time.Time  `json:"created_at"`
	Entries               []Entry    `json:"entries"`
}

type Entry struct {
	EventID              string          `json:"event_id"`
	EntryTypeID          int64           `json:"event_type_id"`
	Amount               decimal.Decimal `json:"amount"`
	DebitAccount         string          `json:"debit_account,omitempty"`
	CreditAccount        string          `json:"credit_account,omitempty"`
	CostCenterDebitOrg   string          `json:"cost_center_debit_org,omitempty"`
	CostCenterDebitCost  string          `json:"cost_center_debit_cost,omitempty"`
	CostCenterCreditOrg  string          `json:"cost_center_credit_org,omitempty"`
	CostCenterCreditCost string          `json:"cost_center_credit_cost,omitempty"`
	CompanyCode          string          `json:"company_code"`
	CompanyType          string          `json:"company_type"`
	Description          string          `json:"description"`
}

func buildEventTable(cid string, e events.Event) (Event, error) {
	evt := Event{
		EventID:         e.EventID,
		CorrelationID:   cid,
		OrgID:           e.OrgID,
		EventTypeID:     e.EventTypeID,
		TransactionID:   e.TransactionID,
		AuthorizationID: e.AuthorizationID,
		ProgramID:       e.ProgramID,
		CustomerID:      e.CustomerID,
		AccountID:       e.AccountID,
		DaysDue:         e.DaysDue,
		EventDate:       e.EventDate,
		Description:     e.Description,
		TransactionType: e.TransactionType,
		ProcessID:       e.ProcessID,
		Producer:        e.Producer,
		ClearingDate:    e.ClearingDate,
		ChunkPart:       e.ChunkPart,
		CreatedAt:       e.CreatedAt,
		Entries:         make([]Entry, 0, len(e.Entries)),
	}

	if e.Processing != nil {
		evt.ProcessingCode = e.Processing.Code
		evt.ProcessingDescription = e.Processing.Description
	}

	for _, en := range e.Entries {
		ent, err := buildEntryTable(e.EventID, en)
		if err != nil {
			return Event{}, err
		}
		evt.Entries = append(evt.Entries, ent)
	}

	return evt, nil
}

func buildEntryTable(eventID string, e events.Entry) (Entry, error) {
	ent := Entry{
		EventID:       eventID,
		EntryTypeID:   e.EntryTypeID,
		Amount:        e.Amount,
		DebitAccount:  e.DebitAccount,
		CreditAccount: e.CreditAccount,
		CompanyCode:   e.Company.Code,
		CompanyType:   e.Company.Type,
		Description:   e.Description,
	}

	if e.CostCenter != nil {
		ent.CostCenterDebitOrg = e.CostCenter.DebitOrg
		ent.CostCenterDebitCost = e.CostCenter.DebitCost
		ent.CostCenterCreditOrg = e.CostCenter.CreditOrg
		ent.CostCenterCreditCost = e.CostCenter.CreditCost
	}

	return ent, nil
}
