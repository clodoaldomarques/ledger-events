package configs

type Flow = string

const (
	Regular   Flow = "regular"
	Migration Flow = "migration"
)

type Script struct {
	ScriptID      int64
	Flow          Flow
	Description   string
	Expression    string
	DebitAccount  *Account
	CreditAccount *Account
}
