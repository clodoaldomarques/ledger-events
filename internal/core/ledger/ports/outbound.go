package ports

import (
	"context"

	"github.com/clodoaldomarques/ledger-events/internal/core/ledger/domain"
)

//go:generate mockgen -source=outbound.go -destination=mock.go -package=ports
type EventRepository interface {
	SaveEvent(ctx context.Context, cid string, e domain.Event) error
}

type EventConfigProvider interface {
	FindConfigByLevel(ctx context.Context, cid string, processing_code string, orgID string, programID int64) (domain.Config, error)
}

type EventPublisher interface {
	Emit(ctx context.Context, cid string, e domain.Event) error
}

type HealthChecker interface {
	Health(ctx context.Context) error
}
