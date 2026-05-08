package events

import (
	"context"

	"github.com/shopspring/decimal"
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
	c, err := s.api.FindConfigByLevel(ctx, cid, e.ProcessingCode, e.OrgID, e.ProgramID)
	if err != nil {
		return Event{}, err
	}

	if err := e.Process(c, amount, fee); err != nil {
		return Event{}, err
	}

	if err := e.Validate(); err != nil {
		return Event{}, err
	}

	if err := s.rep.SaveEvent(ctx, cid, e); err != nil {
		return Event{}, err
	}

	if err := s.top.Emit(ctx, cid, e); err != nil {
		return Event{}, err
	}

	return e, nil
}
