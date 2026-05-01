package scripts

type Flow = string

const (
	Regular   Flow = "regular"
	Migration Flow = "migration"
)

type Entry struct {
	EntryTypeID   int64
	Flow          Flow
	Description   string
	AmountName    string
	Expression    string
	CashInBucket  string
	CashOutBucket string
	CostCenter    *CostCenter
	DebitAccount  *Account
	CreditAccount *Account
	Parameter     *Parameter
}

var legacyExp = map[string]string{
	"amount":          "Amount.amount",
	"amount_interest": "Amount.amount + Fee.iof",
	"amount_duodate":  "Amount.amount + Fee.duodate",
	"interest":        "Amount.interest",
}

func (e Entry) RetrieveExpr() string {
	if e.Expression != "" {
		return e.Expression
	}
	if _, has := legacyExp[e.AmountName]; has {
		return legacyExp[e.AmountName]
	}
	return "Amount.amount"
}
