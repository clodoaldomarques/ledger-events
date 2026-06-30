package events

import (
	"fmt"
	"time"
)

type Event struct {
	EventID        string
	OrgID          string
	ProcessingCode string
	ProgramID      int64
	AccountID      int64
	Description    string
	Producer       string
	CreatedAt      time.Time
	Entries        []Entry
}

func (e Event) Validate() error {
	if len(e.Entries) == 0 {
		return ErrEntryNotFound{fmt.Sprintf("entry not found to producer %s", e.Producer)}
	}
	return nil
}
