package events

import (
	"context"

	"github.com/clodoaldomarques/core-sdk/pkg/otel/tracer"
	"github.com/clodoaldomarques/ledger-events/internal/domain/events"
	"github.com/shopspring/decimal"
)

type Service struct {
	api ConfigProvider
	rep Repository
	top Topic
}

func New(a ConfigProvider, r Repository, t Topic) *Service {
	return &Service{
		rep: r,
		api: a,
		top: t,
	}
}

var p = map[string]func(events.Config, *events.Event, map[string]decimal.Decimal, map[string]decimal.Decimal) error{
	events.Regular:   events.ProcessRegular,
	events.Migration: events.ProcessMigration,
}

func (s Service) CreateEvent(ctx context.Context, cid string, e events.Event, a, f map[string]decimal.Decimal) (events.Event, error) {
	span, ctx := tracer.NewSpanFromContext(ctx, "Service::CreateEvent", map[string]any{
		"cid":   cid,
		"event": e,
	})
	defer span.End()

	c, err := s.api.FindConfigByLevel(ctx, cid, e.ProcessingCode, e.OrgID, e.ProgramID)
	if err != nil {
		span.AddAttributes(map[string]any{
			"account_id": e.AccountID,
			"event":      e,
		})
		span.SetError(err)
		return events.Event{}, err
	}

	if err := p[e.Producer](c, &e, a, f); err != nil {
		span.AddAttributes(map[string]any{
			"account_id": e.AccountID,
			"event":      e,
		})
		span.SetError(err)
		return events.Event{}, err
	}

	if err := e.Validate(); err != nil {
		span.AddAttributes(map[string]any{
			"account_id": e.AccountID,
			"event":      e,
		})
		span.SetError(err)
		return events.Event{}, err
	}

	if err := s.rep.SaveEvent(ctx, cid, e); err != nil {
		span.AddAttributes(map[string]any{
			"account_id": e.AccountID,
			"event":      e,
		})
		span.SetError(err)
		return events.Event{}, err
	}

	if err := s.top.Emit(ctx, cid, e); err != nil {
		span.AddAttributes(map[string]any{
			"account_id": e.AccountID,
			"event":      e,
		})
		span.SetError(err)
		return events.Event{}, err
	}

	return e, nil
}
