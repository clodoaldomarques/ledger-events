package api

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/clodoaldomarques/core-sdk/pkg/logger"
	"github.com/clodoaldomarques/ledger-events/config"
	"github.com/clodoaldomarques/ledger-events/internal/domain/configs"
	"github.com/sony/gobreaker"
)

type LedgerConfigApi struct {
	baseUrl        string
	httpClient     *http.Client
	circuitBreaker *gobreaker.CircuitBreaker
}

func New(ctx context.Context) *LedgerConfigApi {
	baseUrl := config.New().LedgerConfigApiUrl
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
		httpClient:     &http.Client{},
		circuitBreaker: cb,
	}
}

func (a LedgerConfigApi) FindConfigByLevel(ctx context.Context, cid string, processing_code string, orgID string, programID int64) (configs.Config, error) {
	response, err := a.circuitBreaker.Execute(func() (interface{}, error) {
		url := fmt.Sprintf("%s/v1/ledger/config/%s/%d", a.baseUrl, processing_code, programID)
		req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
		if err != nil {
			return configs.Config{}, err
		}
		req.Header.Add("Content-Type", "application/json")
		req.Header.Add("x-cid", cid)
		req.Header.Add("x-tenant", orgID)

		resp, err := a.httpClient.Do(req)
		if err != nil {
			return configs.Config{}, err
		}
		defer resp.Body.Close()

		if resp.StatusCode == http.StatusNotFound {
			body, err := io.ReadAll(resp.Body)
			if err != nil {
				return configs.Config{}, fmt.Errorf("error on read response: %w", err)
			}

			var errResp ErrResponse
			if err := json.Unmarshal(body, &errResp); err != nil {
				return configs.Config{}, fmt.Errorf("unmarshal error: %s", err.Error())
			}
			return configs.Config{}, errResp
		}

		if resp.StatusCode != http.StatusOK {
			return configs.Config{}, fmt.Errorf("api error: status %d", resp.StatusCode)
		}

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return configs.Config{}, fmt.Errorf("erro on read response: %w", err)
		}

		var scriptResponse ConfigResponse
		if err := json.Unmarshal(body, &scriptResponse); err != nil {
			return configs.Config{}, fmt.Errorf("unmarshal erro: %s", err.Error())
		}

		return scriptResponse.ToEntity(), nil

	})

	if err != nil {
		return configs.Config{}, err
	}

	return response.(configs.Config), nil
}

func (a LedgerConfigApi) Close() {
	a.httpClient.CloseIdleConnections()
}
