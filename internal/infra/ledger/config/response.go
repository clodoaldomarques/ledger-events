package config

import (
	"time"

	"github.com/clodoaldomarques/ledger-events/internal/domain/events"
)

type ConfigResponse struct {
	ConfigID    string           `json:"config_id"`
	Level       string           `json:"level"`
	ProcessCode string           `json:"process_code"`
	OrgID       string           `json:"org_id"`
	ProgramID   *int64           `json:"program_id,omitempty"`
	Description string           `json:"description"`
	Scripts     []ScriptResponse `json:"scripts"`
	Enable      bool             `json:"enable"`
	CreatedAt   time.Time        `json:"created_at"`
	UpdatedAt   time.Time        `json:"updated_at"`
	Version     int64            `json:"version"`
}

func (c ConfigResponse) RetrieveDescription() string {
	return c.Description
}

func (c ConfigResponse) RetrieveEntryByProducer(producer string) []events.Script {
	var entries []events.Script

	for _, e := range c.Scripts {
		if e.Flow == producer {
			entries = append(entries, e)
		}
	}

	return entries
}

type AccountResponse struct {
	Number      string `json:"number"`
	Description string `json:"description"`
	Cosif       string `json:"cosif,omitempty"`
}

type ScriptResponse struct {
	ScriptID      int64            `json:"script_id"`
	Flow          string           `json:"flow"`
	Description   string           `json:"description"`
	Expression    string           `json:"expression,omitempty"`
	DebitAccount  *AccountResponse `json:"debit_account,omitempty"`
	CreditAccount *AccountResponse `json:"credit_account,omitempty"`
}

func (s ScriptResponse) RetrieveScriptID() int64 {
	return s.ScriptID
}

func (s ScriptResponse) RetrieveDescription() string {
	return s.Description
}

func (s ScriptResponse) RetrieveExpression() string {
	return s.Expression
}

func (s ScriptResponse) RetrieveCreditAccount() string {
	if s.CreditAccount != nil {
		return s.CreditAccount.Number
	}
	return ""
}

func (s ScriptResponse) RetrieveDebitAccount() string {
	if s.DebitAccount != nil {
		return s.DebitAccount.Number
	}
	return ""
}
