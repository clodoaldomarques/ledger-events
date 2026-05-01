package events

import (
	"context"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"

	"github.com/clodoaldomarques/ledger-events/internal/domain/scripts"
	"github.com/stretchr/testify/assert"
)

func TestService_CreateEvent(t *testing.T) {
	tests := []struct {
		name  string
		setup func(ctrl *gomock.Controller) *Service
		args  func() (Event, map[string]decimal.Decimal, map[string]decimal.Decimal)
		want  func(t *testing.T, evt Event, e error)
	}{
		{
			name: "when calculate various values and save event with success",
			setup: func(ctrl *gomock.Controller) *Service {
				a := NewMockApi(ctrl)
				scr := fakeScript("amount_interest", "(Amount.amount + Fee.iof) * (Fee.tax / Fee.iof) - Fee.tax")
				a.EXPECT().FindScriptByLevel(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(scr, nil).Times(1)

				r := NewMockRepository(ctrl)
				r.EXPECT().SaveEvent(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).Times(1)

				t := NewMockTopic(ctrl)
				t.EXPECT().Emit(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).Times(1)

				return New(a, r, t)
			},
			args: func() (Event, map[string]decimal.Decimal, map[string]decimal.Decimal) {
				return FakeEvent(), FakeAmount(), FakeFee()
			},
			want: func(t *testing.T, evt Event, e error) {
				assert.Nil(t, e)
				got := decimal.NewFromFloat(150.00)
				assert.Equal(t, got.String(), evt.Entries[0].Amount.String())
			},
		},
		{
			name: "when pass single value and save event with success",
			setup: func(ctrl *gomock.Controller) *Service {
				a := NewMockApi(ctrl)
				scr := fakeScript("amount_interest", "Amount.amount")
				a.EXPECT().FindScriptByLevel(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(scr, nil).Times(1)

				r := NewMockRepository(ctrl)
				r.EXPECT().SaveEvent(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).Times(1)

				t := NewMockTopic(ctrl)
				t.EXPECT().Emit(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).Times(1)

				return New(a, r, t)
			},
			args: func() (Event, map[string]decimal.Decimal, map[string]decimal.Decimal) {
				return FakeEvent(), FakeAmount(), FakeFee()
			},
			want: func(t *testing.T, evt Event, e error) {
				assert.Nil(t, e)
				got := decimal.NewFromFloat(150.00)
				assert.Equal(t, got.String(), evt.Entries[0].Amount.String())
			},
		},
		{
			name: "when calculate passing legacy expression and save event with success",
			setup: func(ctrl *gomock.Controller) *Service {
				a := NewMockApi(ctrl)
				scr := fakeScript("amount_interest", "")
				a.EXPECT().FindScriptByLevel(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(scr, nil).Times(1)
				r := NewMockRepository(ctrl)
				r.EXPECT().SaveEvent(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).Times(1)

				t := NewMockTopic(ctrl)
				t.EXPECT().Emit(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).Times(1)

				return New(a, r, t)
			},
			args: func() (Event, map[string]decimal.Decimal, map[string]decimal.Decimal) {
				return FakeEvent(), FakeAmount(), FakeFee()
			},
			want: func(t *testing.T, evt Event, e error) {
				assert.Nil(t, e)
				got := decimal.NewFromFloat(160.00)
				assert.Equal(t, got.String(), evt.Entries[0].Amount.String())
			},
		},
		{
			name: "when receive error on validate expression",
			setup: func(ctrl *gomock.Controller) *Service {
				a := NewMockApi(ctrl)
				scr := fakeScript("amount_interest", "jack sparrow is here")
				a.EXPECT().FindScriptByLevel(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(scr, nil).Times(1)
				r := NewMockRepository(ctrl)
				t := NewMockTopic(ctrl)
				return New(a, r, t)
			},
			args: func() (Event, map[string]decimal.Decimal, map[string]decimal.Decimal) {
				return FakeEvent(), FakeAmount(), FakeFee()
			},
			want: func(t *testing.T, evt Event, e error) {
				assert.NotNil(t, e)
				assert.Equal(t, e.Error(), "invalid expression: jack sparrow is here")
			},
		},
		{
			name: "when script not found",
			setup: func(ctrl *gomock.Controller) *Service {
				a := NewMockApi(ctrl)
				a.EXPECT().FindScriptByLevel(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(scripts.Script{}, scripts.ErrScriptNotFound{}).Times(1)
				r := NewMockRepository(ctrl)
				t := NewMockTopic(ctrl)

				return New(a, r, t)
			},
			args: func() (Event, map[string]decimal.Decimal, map[string]decimal.Decimal) {
				return FakeEvent(), FakeAmount(), FakeFee()
			},
			want: func(t *testing.T, evt Event, e error) {
				assert.NotNil(t, e)
				assert.Equal(t, e.Error(), "accounting script not found")
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

func fakeScript(amountName string, expr string) scripts.Script {
	return scripts.Script{
		ScriptID:    "PISMO-B612",
		Level:       scripts.PlatformLevel,
		EventTypeID: "b-612",
		OrgID:       "TN-123-456",
		Description: "programa de teste",
		Entries: []scripts.Entry{
			{
				EntryTypeID: 100,
				Flow:        scripts.Regular,
				Description: "entrada de teste",
				AmountName:  amountName,
				Expression:  expr,
			},
			{
				EntryTypeID: 101,
				Flow:        scripts.Migration,
				Description: "entrada de teste",
				AmountName:  amountName,
				Expression:  expr,
			},
		},
	}
}

func FakeEvent() Event {
	transID := new(int64)
	return Event{
		EventID:       uuid.NewString(),
		TransactionID: transID,
		OrgID:         "TN-123-456",
		EventTypeID:   "b-612",
		ProgramID:     400,
		Producer:      "migration",
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
