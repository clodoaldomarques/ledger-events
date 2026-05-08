package api

import (
	"time"

	"github.com/clodoaldomarques/ledger-events/internal/domain/configs"
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

func (s ConfigResponse) ToEntity() configs.Config {
	scr := configs.Config{
		ConfigID:    s.ConfigID,
		Level:       configs.Level(s.Level),
		ProcessCode: s.ProcessCode,
		OrgID:       s.OrgID,
		Description: s.Description,
		Scripts:     make([]configs.Script, 0, len(s.Scripts)),
		CreatedAt:   s.CreatedAt,
		UpdatedAt:   s.UpdatedAt,
		Enable:      s.Enable,
		Version:     s.Version,
	}

	if s.ProgramID != nil {
		scr.ProgramID = *s.ProgramID
	}

	for _, e := range s.Scripts {
		scr.Scripts = append(scr.Scripts, e.ToEntity())
	}

	return scr
}

type AccountResponse struct {
	Number      string `json:"number"`
	Description string `json:"description"`
	Cosif       string `json:"cosif,omitempty"`
}

func (a AccountResponse) ToEntity() *configs.Account {
	return &configs.Account{
		Number:      a.Number,
		Description: a.Description,
		Cosif:       a.Cosif,
	}
}

type ScriptResponse struct {
	ScriptID      int64            `json:"script_id"`
	Flow          string           `json:"flow"`
	Description   string           `json:"description"`
	Expression    string           `json:"expression,omitempty"`
	DebitAccount  *AccountResponse `json:"debit_account,omitempty"`
	CreditAccount *AccountResponse `json:"credit_account,omitempty"`
}

func (e ScriptResponse) ToEntity() configs.Script {
	ent := configs.Script{
		ScriptID:    e.ScriptID,
		Flow:        e.Flow,
		Description: e.Description,
		Expression:  e.Expression,
	}

	if e.DebitAccount != nil {
		ent.DebitAccount = e.DebitAccount.ToEntity()
	}

	if e.CreditAccount != nil {
		ent.CreditAccount = e.CreditAccount.ToEntity()
	}

	return ent
}
