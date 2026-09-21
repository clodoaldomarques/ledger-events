package events

import (
	"context"

	"github.com/clodoaldomarques/ledger-events/internal/domain/events"
)

//go:generate mockgen -source=interfaces.go -destination=mock.go -package=events
type Repository interface {
	SaveEvent(ctx context.Context, cid string, e events.Event) error
}

type ConfigProvider interface {
	FindConfigByLevel(ctx context.Context, cid string, processing_code string, orgID string, programID int64) (events.Config, error)
}

type Topic interface {
	Emit(ctx context.Context, cid string, e events.Event) error
}
