package accounting

import (
	"time"

	"github.com/clodoaldomarques/ledger-events/internal/domain/events"
	"github.com/shopspring/decimal"
)

type EventResponse struct {
	EventID         string              `json:"event_id"`
	OrgID           string              `json:"org_id"`
	EventTypeID     string              `json:"event_type_id"`
	TransactionID   *int64              `json:"transaction_id,omitempty"`
	AuthorizationID *int64              `json:"authorization_id,omitempty"`
	ProgramID       int64               `json:"program_id"`
	CustomerID      *int64              `json:"customer_id,omitempty"`
	AccountID       int64               `json:"account_id"`
	DaysDue         int64               `json:"days_due,omitempty"`
	EventDate       time.Time           `json:"event_date"`
	Description     string              `json:"description"`
	TransactionType int64               `json:"transaction_type"`
	ProcessID       int64               `json:"process_id"`
	Producer        string              `json:"producer"`
	Processing      *ProcessingResponse `json:"processing,omitempty"`
	ClearingDate    *time.Time          `json:"clearing_date,omitempty"`
	ChunkPart       int64               `json:"chunk_part"`
	CreatedAt       time.Time           `json:"created_at"`
	Entries         []EntryResponse     `json:"entries"`
}

func buildEventResponse(e events.Event) EventResponse {
	er := EventResponse{
		EventID:         e.EventID,
		OrgID:           e.OrgID,
		EventTypeID:     e.EventTypeID,
		TransactionID:   e.TransactionID,
		AuthorizationID: e.AuthorizationID,
		ProgramID:       e.ProcessID,
		CustomerID:      e.CustomerID,
		AccountID:       e.AccountID,
		DaysDue:         e.DaysDue,
		Description:     e.Description,
		TransactionType: e.TransactionType,
		ProcessID:       e.ProcessID,
		Producer:        e.Producer,
		ChunkPart:       e.ChunkPart,
		CreatedAt:       e.CreatedAt,
	}

	if e.Processing != nil {
		er.Processing = buildProcessingResponse(e.Processing)
	}

	for _, en := range e.Entries {
		er.Entries = append(er.Entries, buildEntryResponse(en))
	}

	return er
}

type ProcessingResponse struct {
	Code        string `json:"code"`
	Description string `json:"description"`
}

func buildProcessingResponse(p *events.Processing) *ProcessingResponse {
	return &ProcessingResponse{
		Code:        p.Code,
		Description: p.Description,
	}
}

type EntryResponse struct {
	EntryTypeID   int64               `json:"entry_type_id"`
	Amount        decimal.Decimal     `json:"amount"`
	DebitAccount  string              `json:"debit_account"`
	CreditAccount string              `json:"credit_account"`
	CostCenter    *CostCenterResponse `json:"cost_center,omitempty"`
	Company       CompanyResponse     `json:"company"`
	Description   string              `json:"description"`
}

func buildEntryResponse(e events.Entry) EntryResponse {
	er := EntryResponse{
		EntryTypeID:   e.EntryTypeID,
		Amount:        e.Amount,
		DebitAccount:  e.DebitAccount,
		CreditAccount: e.CreditAccount,
		Company:       buildCompanyResponse(e.Company),
		Description:   e.Description,
	}

	if e.CostCenter != nil {
		er.CostCenter = buildCostCenterResponse(e.CostCenter)
	}

	return er
}

type CompanyResponse struct {
	Code string `json:"code"`
	Type string `json:"type"`
}

func buildCompanyResponse(c events.Company) CompanyResponse {
	return CompanyResponse{
		Code: c.Code,
		Type: c.Type,
	}
}

type CostCenterResponse struct {
	DebitOrg   string `json:"debit_org"`
	DebitCost  string `json:"debit_cost"`
	CreditOrg  string `json:"credit_org"`
	CreditCost string `json:"credit_cost"`
}

func buildCostCenterResponse(c *events.CostCenter) *CostCenterResponse {
	return &CostCenterResponse{
		DebitOrg:   c.DebitOrg,
		DebitCost:  c.DebitCost,
		CreditOrg:  c.CreditOrg,
		CreditCost: c.CreditCost,
	}
}
