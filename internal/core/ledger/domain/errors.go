package domain

type ErrScriptNotFound struct {
}

func (e ErrScriptNotFound) Error() string {
	return "ledger config not found"
}

type ErrEntryNotFound struct {
	msg string
}

func (e ErrEntryNotFound) Error() string {
	if e.msg == "" {
		return "entry no found"
	}
	return e.msg
}
