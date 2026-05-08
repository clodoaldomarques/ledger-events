package configs

import (
	"time"
)

type Level string

const (
	PlatformLevel Level = "platform"
	TenantLevel   Level = "tenant"
	ProgramLevel  Level = "program"
)

type Config struct {
	ConfigID    string
	Level       Level
	ProcessCode string
	OrgID       string
	ProgramID   int64
	Description string
	Scripts     []Script
	Enable      bool
	CreatedAt   time.Time
	UpdatedAt   time.Time
	Version     int64
}

func (s Config) RetrieveEntryByProducer(producer string) []Script {
	m := make([]Script, 0)
	for _, e := range s.Scripts {
		if string(e.Flow) == producer {
			m = append(m, e)
		}
	}
	return m
}
