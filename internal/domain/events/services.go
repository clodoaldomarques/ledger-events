package events

import (
	"context"

	"github.com/clodoaldomarques/core-sdk/pkg/tracer"
	"github.com/clodoaldomarques/ledger-events/internal/domain/configs"
	"github.com/shopspring/decimal"
	"go.opentelemetry.io/otel/attribute"
)

type Service struct {
	api ConfigApi
	rep Repository
	top Topic
}

func New(a ConfigApi, r Repository, t Topic) *Service {
	return &Service{
		rep: r,
		api: a,
		top: t,
	}
}

var p = map[string]func(configs.Config, *Event, map[string]decimal.Decimal, map[string]decimal.Decimal) error{
	configs.Regular:   ProcessRegular,
	configs.Migration: ProcessMigration,
}

func (s Service) CreateEvent(ctx context.Context, cid string, e Event, a, f map[string]decimal.Decimal) (Event, error) {
	span, ctx := tracer.NewSpanFromContext(ctx, "Service::CreateEvent", attribute.String("cid", cid))
	defer span.End()

	c, err := s.api.FindConfigByLevel(ctx, cid, e.ProcessingCode, e.OrgID, e.ProgramID)
	if err != nil {
		span.AddAttributes(tracer.Attributes{
			"account_id": e.AccountID,
			"event":      e,
		})
		span.SetError(err)
		return Event{}, err
	}

	if err := p[e.Producer](c, &e, a, f); err != nil {
		span.AddAttributes(tracer.Attributes{
			"account_id": e.AccountID,
			"event":      e,
		})
		span.SetError(err)
		return Event{}, err
	}

	if err := e.Validate(); err != nil {
		span.AddAttributes(tracer.Attributes{
			"account_id": e.AccountID,
			"event":      e,
		})
		span.SetError(err)
		return Event{}, err
	}

	if err := s.rep.SaveEvent(ctx, cid, e); err != nil {
		span.AddAttributes(tracer.Attributes{
			"account_id": e.AccountID,
			"event":      e,
		})
		span.SetError(err)
		return Event{}, err
	}

	if err := s.top.Emit(ctx, cid, e); err != nil {
		span.AddAttributes(tracer.Attributes{
			"account_id": e.AccountID,
			"event":      e,
		})
		span.SetError(err)
		return Event{}, err
	}

	return e, nil
}
