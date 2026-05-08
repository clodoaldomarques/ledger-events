package events

import (
	"context"

	"github.com/clodoaldomarques/ledger-events/internal/domain/configs"
)

//go:generate mockgen -source=interfaces.go -destination=mock.go -package=events
type Repository interface {
	SaveEvent(ctx context.Context, cid string, e Event) error
}

type Api interface {
	FindConfigByLevel(ctx context.Context, cid string, processing_code string, orgID string, programID int64) (configs.Config, error)
}

type Topic interface {
	Emit(ctx context.Context, cid string, e Event) error
}
