package api

import (
	"time"

	"github.com/clodoaldomarques/ledger-events/internal/domain/scripts"
)

type ScriptResponse struct {
	ScriptID    string           `json:"script_id"`
	Level       string           `json:"level"`
	EventTypeID string           `json:"event_type_id"`
	OrgID       string           `json:"org_id"`
	ProgramID   *int64           `json:"program_id,omitempty"`
	Description string           `json:"description"`
	Company     *CompanyResponse `json:"company,omitempty"`
	Entries     []EntryResponse  `json:"entries"`
	Enable      bool             `json:"enable"`
	CreatedAt   time.Time        `json:"created_at"`
	UpdatedAt   time.Time        `json:"updated_at"`
	Version     int64            `json:"version"`
}

func (s ScriptResponse) ToEntity() scripts.Script {
	scr := scripts.Script{
		ScriptID:    s.ScriptID,
		Level:       scripts.Level(s.Level),
		EventTypeID: s.EventTypeID,
		OrgID:       s.OrgID,
		Description: s.Description,
		Entries:     make([]scripts.Entry, 0, len(s.Entries)),
		CreatedAt:   s.CreatedAt,
		UpdatedAt:   s.UpdatedAt,
		Enable:      s.Enable,
		Version:     s.Version,
	}

	if s.ProgramID != nil {
		scr.ProgramID = *s.ProgramID
	}

	if s.Company != nil {
		scr.Company = s.Company.ToEntity()
	}

	for _, e := range s.Entries {
		scr.Entries = append(scr.Entries, e.ToEntity())
	}

	return scr
}

type CompanyResponse struct {
	Code string `json:"code,omitempty"`
	Type string `json:"type,omitempty"`
}

func (c CompanyResponse) ToEntity() *scripts.Company {
	return &scripts.Company{
		Code: c.Code,
		Type: c.Type,
	}
}

type AccountResponse struct {
	Number      string `json:"number"`
	Description string `json:"description"`
	Cosif       string `json:"cosif,omitempty"`
}

func (a AccountResponse) ToEntity() *scripts.Account {
	return &scripts.Account{
		Number:      a.Number,
		Description: a.Description,
		Cosif:       a.Cosif,
	}
}

type CostCenterResponse struct {
	DebitCost  string `json:"debit_cost"`
	DebitOrg   string `json:"debit_org"`
	CreditCost string `json:"credit_cost"`
	CreditOrg  string `json:"credit_org"`
}

func (c CostCenterResponse) ToEntity() *scripts.CostCenter {
	return &scripts.CostCenter{
		DebitCost:  c.DebitCost,
		DebitOrg:   c.DebitOrg,
		CreditCost: c.CreditCost,
		CreditOrg:  c.CreditOrg,
	}
}

type EntryResponse struct {
	EntryTypeID   int64               `json:"entry_type_id"`
	Flow          string              `json:"flow"`
	Description   string              `json:"description"`
	AmountName    string              `json:"amount_name,omitempty"`
	Expression    string              `json:"expression,omitempty"`
	CashInBucket  string              `json:"cashin_bucket,omitempty"`
	CashOutBucket string              `json:"cashout_bucket,omitempty"`
	CostCenter    *CostCenterResponse `json:"cost_center,omitempty"`
	DebitAccount  *AccountResponse    `json:"debit_account,omitempty"`
	CreditAccount *AccountResponse    `json:"credit_account,omitempty"`
	Parameter     *ParameterResponse  `json:"parameter,omitempty"`
}

func (e EntryResponse) ToEntity() scripts.Entry {
	ent := scripts.Entry{
		EntryTypeID:   e.EntryTypeID,
		Flow:          e.Flow,
		Description:   e.Description,
		AmountName:    e.AmountName,
		Expression:    e.Expression,
		CashInBucket:  e.CashInBucket,
		CashOutBucket: e.CashOutBucket,
	}

	if e.CostCenter != nil {
		ent.CostCenter = e.CostCenter.ToEntity()
	}

	if e.DebitAccount != nil {
		ent.DebitAccount = e.DebitAccount.ToEntity()
	}

	if e.CreditAccount != nil {
		ent.CreditAccount = e.CreditAccount.ToEntity()
	}

	if e.Parameter != nil {
		ent.Parameter = e.Parameter.ToEntity()
	}
	return ent
}

type ParameterResponse struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

func (p ParameterResponse) ToEntity() *scripts.Parameter {
	return &scripts.Parameter{
		Name:  p.Name,
		Value: p.Value,
	}
}
