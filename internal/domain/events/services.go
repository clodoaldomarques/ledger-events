package events

import (
	"context"

	"github.com/clodoaldomarques/core-sdk/pkg/tracer"
	"github.com/shopspring/decimal"
	"go.opentelemetry.io/otel/attribute"
)

type Service struct {
	api Api
	rep Repository
	top Topic
}

func New(a Api, r Repository, t Topic) *Service {
	return &Service{
		rep: r,
		api: a,
		top: t,
	}
}

func (s Service) CreateEvent(ctx context.Context, cid string, e Event, amount, fee map[string]decimal.Decimal) (Event, error) {
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

	if err := e.Process(c, amount, fee); err != nil {
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
