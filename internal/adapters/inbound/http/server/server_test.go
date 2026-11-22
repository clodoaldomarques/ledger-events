package server

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/clodoaldomarques/ledger-events/internal/adapters/inbound/http/ledger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type fakeHealth struct {
	err      error
	calls    int
	deadline time.Time
}

func (f *fakeHealth) Health(ctx context.Context) error {
	f.calls++
	f.deadline, _ = ctx.Deadline()
	return f.err
}

func newTestServer(health *fakeHealth) *Server {
	return New(ledger.NewHandler(nil), health)
}

func TestReadiness(t *testing.T) {
	health := &fakeHealth{}
	s := newTestServer(health)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/health/ready", nil)
	s.http.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, 1, health.calls)
	assert.WithinDuration(t, time.Now().Add(500*time.Millisecond), health.deadline, 100*time.Millisecond)
}

func TestReadiness_Unavailable(t *testing.T) {
	health := &fakeHealth{err: errors.New("mysql down")}
	s := newTestServer(health)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/health/ready", nil)
	s.http.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusServiceUnavailable, rec.Code)
	assert.Equal(t, 1, health.calls)
	assert.Empty(t, rec.Body.String())
}

func TestLiveness(t *testing.T) {
	s := newTestServer(&fakeHealth{})

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/health/live", nil)
	s.http.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Empty(t, rec.Body.String())
}

func TestNew_ServerImplementsEchoHandler(t *testing.T) {
	var _ http.Handler = newTestServer(&fakeHealth{}).http
}
