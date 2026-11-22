package ledger

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/clodoaldomarques/ledger-events/internal/core/ledger/domain"
	"github.com/labstack/echo/v4"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type fakeUseCase struct {
	calls   int
	cid     string
	event   domain.Event
	amounts map[string]decimal.Decimal
	fees    map[string]decimal.Decimal
	err     error
}

func (f *fakeUseCase) CreateEvent(_ context.Context, cid string, e domain.Event, amounts, fees map[string]decimal.Decimal) (domain.Event, error) {
	f.calls++
	f.cid = cid
	f.event = e
	f.amounts = amounts
	f.fees = fees
	if f.err != nil {
		return domain.Event{}, f.err
	}
	return e, nil
}

func newTestServer(uc *fakeUseCase) *echo.Echo {
	e := echo.New()
	e.POST("/v1/ledger/events", NewHandler(uc).CreateEvent)
	return e
}

const validBody = `{
	"processing_code": "b-612",
	"program_id": 400,
	"account_id": 123,
	"producer": "regular",
	"amounts": {"amount": 150},
	"fees": {"iof": 15}
}`

func TestCreateEvent(t *testing.T) {
	uc := &fakeUseCase{}
	e := newTestServer(uc)

	req := httptest.NewRequest(http.MethodPost, "/v1/ledger/events", strings.NewReader(validBody))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	req.Header.Set("x-tenant", "TN-123")
	req.Header.Set("x-cid", "cid-fixo")
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	require.Equal(t, http.StatusCreated, rec.Code)
	assert.Equal(t, 1, uc.calls)
	assert.Equal(t, "cid-fixo", uc.cid)
	assert.Equal(t, "TN-123", uc.event.OrgID)
	assert.Equal(t, "b-612", uc.event.ProcessingCode)
	assert.Equal(t, int64(400), uc.event.ProgramID)
	assert.Equal(t, int64(123), uc.event.AccountID)
	assert.Equal(t, domain.Regular, uc.event.Producer)
	assert.Equal(t, "150", uc.amounts["amount"].String())
	assert.Equal(t, "15", uc.fees["iof"].String())
	assert.NotEmpty(t, uc.event.EventID)
	assert.False(t, uc.event.CreatedAt.IsZero())

	var response map[string]any
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &response))
	assert.Equal(t, uc.event.EventID, response["event_id"])
	assert.Equal(t, "TN-123", response["org_id"])
}

func TestCreateEvent_GeneratesCID(t *testing.T) {
	uc := &fakeUseCase{}
	e := newTestServer(uc)

	req := httptest.NewRequest(http.MethodPost, "/v1/ledger/events", strings.NewReader(validBody))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	require.Equal(t, http.StatusCreated, rec.Code)
	assert.NotEmpty(t, uc.cid)
}

func TestCreateEvent_UseCaseError(t *testing.T) {
	uc := &fakeUseCase{err: domain.ErrScriptNotFound{}}
	e := newTestServer(uc)

	req := httptest.NewRequest(http.MethodPost, "/v1/ledger/events", strings.NewReader(validBody))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
	assert.JSONEq(t, `{"message":"ledger config not found"}`, rec.Body.String())
}

func TestCreateEvent_InvalidBody(t *testing.T) {
	uc := &fakeUseCase{}
	e := newTestServer(uc)

	req := httptest.NewRequest(http.MethodPost, "/v1/ledger/events", strings.NewReader(`{"processing_code":`))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
	assert.Equal(t, 0, uc.calls)
}
