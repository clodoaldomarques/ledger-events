package events

import (
	"context"

	"github.com/clodoaldomarques/ledger-events/internal/domain/scripts"
)

//go:generate mockgen -source=interfaces.go -destination=mock.go -package=events
type Repository interface {
	SaveEvent(ctx context.Context, cid string, e Event) error
}

type Api interface {
	FindScriptByLevel(ctx context.Context, cid string, eventTypeID string, orgID string, programID int64) (scripts.Script, error)
}

type Topic interface {
	Emit(ctx context.Context, cid string, e Event) error
}
