package events

import (
	"context"
	"testing"

	"github.com/clodoaldomarques/ledger-events/internal/domain/events"
	"github.com/clodoaldomarques/ledger-events/internal/infra/ledger/config"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"go.uber.org/mock/gomock"

	"github.com/stretchr/testify/assert"
)

func TestService_CreateEvent(t *testing.T) {
	tests := []struct {
		name  string
		setup func(ctrl *gomock.Controller) *Service
		args  func() (events.Event, map[string]decimal.Decimal, map[string]decimal.Decimal)
		want  func(t *testing.T, evt events.Event, e error)
	}{
		{
			name: "when calculate various values and save event with success",
			setup: func(ctrl *gomock.Controller) *Service {
				a := NewMockConfigProvider(ctrl)
				scr := fakeScript("(Amounts.amount + Fees.iof) * (Fees.tax / Fees.iof) - Fees.tax")
				a.EXPECT().FindConfigByLevel(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(scr, nil).Times(1)

				r := NewMockRepository(ctrl)
				r.EXPECT().SaveEvent(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).Times(1)

				t := NewMockTopic(ctrl)
				t.EXPECT().Emit(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).Times(1)

				return New(a, r, t)
			},
			args: func() (events.Event, map[string]decimal.Decimal, map[string]decimal.Decimal) {
				return FakeEvent(), FakeAmount(), FakeFee()
			},
			want: func(t *testing.T, evt events.Event, e error) {
				assert.Nil(t, e)
				got := decimal.NewFromFloat(150.00)
				assert.Equal(t, got.String(), evt.Entries[0].Amount.String())
			},
		},
		{
			name: "when pass single value and save event with success",
			setup: func(ctrl *gomock.Controller) *Service {
				a := NewMockConfigProvider(ctrl)
				scr := fakeScript("Amounts.amount")
				a.EXPECT().FindConfigByLevel(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(scr, nil).Times(1)

				r := NewMockRepository(ctrl)
				r.EXPECT().SaveEvent(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).Times(1)

				t := NewMockTopic(ctrl)
				t.EXPECT().Emit(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).Times(1)

				return New(a, r, t)
			},
			args: func() (events.Event, map[string]decimal.Decimal, map[string]decimal.Decimal) {
				return FakeEvent(), FakeAmount(), FakeFee()
			},
			want: func(t *testing.T, evt events.Event, e error) {
				assert.Nil(t, e)
				got := decimal.NewFromFloat(150.00)
				assert.Equal(t, got.String(), evt.Entries[0].Amount.String())
			},
		},
		{
			name: "when receive error on validate expression",
			setup: func(ctrl *gomock.Controller) *Service {
				a := NewMockConfigProvider(ctrl)
				scr := fakeScript("jack sparrow is here")
				a.EXPECT().FindConfigByLevel(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(scr, nil).Times(1)
				r := NewMockRepository(ctrl)
				t := NewMockTopic(ctrl)
				return New(a, r, t)
			},
			args: func() (events.Event, map[string]decimal.Decimal, map[string]decimal.Decimal) {
				return FakeEvent(), FakeAmount(), FakeFee()
			},
			want: func(t *testing.T, evt events.Event, e error) {
				assert.NotNil(t, e)
				assert.Equal(t, e.Error(), "invalid expression: jack sparrow is here")
			},
		},
		{
			name: "when script not found",
			setup: func(ctrl *gomock.Controller) *Service {
				a := NewMockConfigProvider(ctrl)
				a.EXPECT().FindConfigByLevel(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, events.ErrScriptNotFound{}).Times(1)
				r := NewMockRepository(ctrl)
				t := NewMockTopic(ctrl)

				return New(a, r, t)
			},
			args: func() (events.Event, map[string]decimal.Decimal, map[string]decimal.Decimal) {
				return FakeEvent(), FakeAmount(), FakeFee()
			},
			want: func(t *testing.T, evt events.Event, e error) {
				assert.NotNil(t, e)
				assert.Equal(t, e.Error(), "ledger config not found")
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			s := tt.setup(ctrl)
			e, a, f := tt.args()
			evt, err := s.CreateEvent(context.Background(), uuid.NewString(), e, a, f)
			tt.want(t, evt, err)
		})
	}
}

func fakeScript(expr string) events.Config {
	var c any = config.ConfigResponse{
		ConfigID:    "LEDGER-B612",
		Level:       "tenant",
		ProcessCode: "b-612",
		OrgID:       "TN-123-456",
		Description: "programa de teste",
		Scripts: []config.ScriptResponse{
			{
				ScriptID:    100,
				Flow:        "regular",
				Description: "entrada de teste",
				Expression:  expr,
			},
			{
				ScriptID:    101,
				Flow:        "migration",
				Description: "entrada de teste",
				Expression:  expr,
			},
		},
	}

	return c.(events.Config)
}

func FakeEvent() events.Event {
	return events.Event{
		EventID:        uuid.NewString(),
		OrgID:          "TN-123-456",
		ProcessingCode: "b-612",
		ProgramID:      400,
		Producer:       "migration",
	}
}

func FakeAmount() map[string]decimal.Decimal {
	return map[string]decimal.Decimal{
		"amount": decimal.NewFromFloat(150.00),
	}

}

func FakeFee() map[string]decimal.Decimal {
	return map[string]decimal.Decimal{
		"iof": decimal.NewFromFloat(10.00),
		"tax": decimal.NewFromFloat(10.00),
	}
}
