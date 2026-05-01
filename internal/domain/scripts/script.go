package scripts

import (
	"time"
)

type Level string

const (
	PlatformLevel Level = "platform"
	TenantLevel   Level = "tenant"
	ProgramLevel  Level = "program"
)

type Script struct {
	ScriptID    string
	Level       Level
	EventTypeID string
	OrgID       string
	ProgramID   int64
	Description string
	Company     *Company
	Entries     []Entry
	Enable      bool
	CreatedAt   time.Time
	UpdatedAt   time.Time
	Version     int64
}

func (s Script) RetrieveEntryByProducer(producer string) []Entry {
	m := make([]Entry, 0)
	for _, e := range s.Entries {
		switch producer {
		case "migration":
			if e.Flow == Regular || e.Flow == Migration {
				m = append(m, e)
			}
		default:
			if e.Flow == Regular {
				m = append(m, e)
			}
		}
	}
	return m
}
