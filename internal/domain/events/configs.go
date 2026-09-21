package events

const (
	Regular   = "regular"
	Migration = "migration"
)

type Config interface {
	RetrieveDescription() string
	RetrieveEntryByProducer(producer string) []Script
}

type Script interface {
	RetrieveScriptID() int64
	RetrieveDescription() string
	RetrieveExpression() string
	RetrieveCreditAccount() string
	RetrieveDebitAccount() string
}
