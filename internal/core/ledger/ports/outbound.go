package ports

import (
	"context"

	"github.com/clodoaldomarques/ledger-events/internal/core/ledger/domain"
)

//go:generate mockgen -source=outbound.go -destination=mock.go -package=ports
type Repository interface {
	SaveEvent(ctx context.Context, cid string, e domain.Event) error
}

type ConfigProvider interface {
	FindConfigByLevel(ctx context.Context, cid string, processing_code string, orgID string, programID int64) (domain.Config, error)
}

type Topic interface {
	Emit(ctx context.Context, cid string, e domain.Event) error
}
