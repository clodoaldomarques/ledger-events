package application

import (
	"github.com/clodoaldomarques/ledger-events/internal/core/ledger/domain"
)

// stubConfig e stubScript implementam as interfaces de domínio Config/Script sem
// depender do adapter HTTP do Ledger Config.
type stubConfig struct {
	description string
	scripts     []stubScript
}

func (c stubConfig) RetrieveDescription() string {
	return c.description
}

func (c stubConfig) RetrieveEntryByProducer(producer string) []domain.Script {
	entries := make([]domain.Script, 0)
	for _, script := range c.scripts {
		if script.flow == producer {
			entries = append(entries, script)
		}
	}
	return entries
}

type stubScript struct {
	id          int64
	flow        string
	description string
	expression  string
}

func (s stubScript) RetrieveScriptID() int64       { return s.id }
func (s stubScript) RetrieveDescription() string   { return s.description }
func (s stubScript) RetrieveExpression() string    { return s.expression }
func (s stubScript) RetrieveCreditAccount() string { return "" }
func (s stubScript) RetrieveDebitAccount() string  { return "" }
