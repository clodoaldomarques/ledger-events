package config

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/clodoaldomarques/core-sdk/pkg/otel/httpclient"
	"github.com/clodoaldomarques/core-sdk/pkg/otel/tracer"
	"github.com/clodoaldomarques/core-sdk/pkg/zap/logger"
	"github.com/clodoaldomarques/ledger-events/config"
	"github.com/clodoaldomarques/ledger-events/internal/core/ledger/domain"
	"github.com/clodoaldomarques/ledger-events/internal/core/ledger/ports"
	"github.com/sony/gobreaker"
)

type LedgerConfigApi struct {
	baseUrl        string
	httpClient     *http.Client
	circuitBreaker *gobreaker.CircuitBreaker
}

var _ ports.EventConfigProvider = (*LedgerConfigApi)(nil)

// New recebe a URL base e o cliente HTTP já construídos pela raiz de composição.
func New(ctx context.Context, baseUrl string, client *http.Client) *LedgerConfigApi {
	cb := gobreaker.NewCircuitBreaker(gobreaker.Settings{
		Name:    "LedgerConfigAPI",
		Timeout: 15 * time.Second,
		ReadyToTrip: func(counts gobreaker.Counts) bool {
			return counts.ConsecutiveFailures > 5
		},
		OnStateChange: func(name string, from, to gobreaker.State) {
			logger.Error(ctx, fmt.Sprintf("circuitbreaker '%s' changed from %v to %v", name, from, to), logger.Fields{
				"api_name": name,
				"api_url":  baseUrl,
			})
		},
	})

	return &LedgerConfigApi{
		baseUrl:        baseUrl,
		httpClient:     client,
		circuitBreaker: cb,
	}
}

// NewFromConfig mantém a construção por requisição usada antes da spec 005.
// Será removido quando a raiz de composição assumir a criação dos adapters.
func NewFromConfig(ctx context.Context) *LedgerConfigApi {
	cfg := config.New()
	return New(ctx, cfg.LedgerConfigApiUrl, httpclient.Client())
}

func (a LedgerConfigApi) FindConfigByLevel(ctx context.Context, cid string, processing_code string, orgID string, programID int64) (domain.Config, error) {
	span, ctx := tracer.NewSpanFromContext(ctx, "LedgerConfigApi::FindConfigByLevel", map[string]any{
		"cid":             cid,
		"processing_code": processing_code,
		"org_id":          orgID,
		"program_id":      programID,
	})
	defer span.End()

	response, err := a.circuitBreaker.Execute(func() (any, error) {
		url := fmt.Sprintf("%s/v1/ledger/config/%s/%d", a.baseUrl, processing_code, programID)
		req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
		if err != nil {
			span.SetError(err)
			return nil, err
		}
		req.Header.Add("Content-Type", "application/json")
		req.Header.Add("x-cid", cid)
		req.Header.Add("x-tenant", orgID)

		resp, err := a.httpClient.Do(req)
		if err != nil {
			span.SetError(err)
			return nil, err
		}
		defer resp.Body.Close()

		if resp.StatusCode == http.StatusNotFound {
			body, err := io.ReadAll(resp.Body)
			if err != nil {
				span.SetError(err)
				return nil, fmt.Errorf("error on read response: %w", err)
			}

			var errResp ErrResponse
			if err := json.Unmarshal(body, &errResp); err != nil {
				span.SetError(err)
				return nil, fmt.Errorf("unmarshal error: %s", err.Error())
			}
			span.SetError(errResp)
			return nil, errResp
		}

		if resp.StatusCode != http.StatusOK {
			span.SetError(fmt.Errorf("api error: status %d", resp.StatusCode))
			return nil, fmt.Errorf("api error: status %d", resp.StatusCode)
		}

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			span.SetError(err)
			return nil, fmt.Errorf("error on read response: %w", err)
		}

		var configResponse ConfigResponse
		if err := json.Unmarshal(body, &configResponse); err != nil {
			span.SetError(err)
			return nil, fmt.Errorf("unmarshal erro: %s", err.Error())
		}

		return configResponse, nil

	})

	if err != nil {
		span.SetError(err)
		return nil, err
	}

	return response.(domain.Config), nil
}

func (a LedgerConfigApi) Close() {
	a.httpClient.CloseIdleConnections()
}
