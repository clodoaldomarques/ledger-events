package api

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/clodoaldomarques/ledger-events/configs"
	"github.com/clodoaldomarques/ledger-events/internal/domain/scripts"
	"github.com/clodoaldomarques/ledger-events/pkg/logger"
	"github.com/sony/gobreaker"
)

type AccountScriptsApi struct {
	baseUrl        string
	httpClient     *http.Client
	circuitBreaker *gobreaker.CircuitBreaker
}

func New(ctx context.Context) *AccountScriptsApi {
	baseUrl := configs.New().AccountScriptsApiUrl
	cb := gobreaker.NewCircuitBreaker(gobreaker.Settings{
		Name:    "AccountingScriptsAPI",
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

	return &AccountScriptsApi{
		baseUrl:        baseUrl,
		httpClient:     &http.Client{},
		circuitBreaker: cb,
	}
}

func (a AccountScriptsApi) FindScriptByLevel(ctx context.Context, cid string, eventTypeID string, orgID string, programID int64) (scripts.Script, error) {
	response, err := a.circuitBreaker.Execute(func() (interface{}, error) {
		url := fmt.Sprintf("%s/v1/accounting/scripts/%s/%d", a.baseUrl, eventTypeID, programID)
		req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
		if err != nil {
			return scripts.Script{}, err
		}
		req.Header.Add("Content-Type", "application/json")
		req.Header.Add("x-cid", cid)
		req.Header.Add("x-tenant", orgID)

		resp, err := a.httpClient.Do(req)
		if err != nil {
			return scripts.Script{}, err
		}
		defer resp.Body.Close()

		if resp.StatusCode == http.StatusNotFound {
			body, err := io.ReadAll(resp.Body)
			if err != nil {
				return scripts.Script{}, fmt.Errorf("error on read response: %w", err)
			}

			var errResp ErrResponse
			if err := json.Unmarshal(body, &errResp); err != nil {
				return scripts.Script{}, fmt.Errorf("unmarshal error: %s", err.Error())
			}
			return scripts.Script{}, errResp
		}

		if resp.StatusCode != http.StatusOK {
			return scripts.Script{}, fmt.Errorf("api error: status %d", resp.StatusCode)
		}

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return scripts.Script{}, fmt.Errorf("erro on read response: %w", err)
		}

		var scriptResponse ScriptResponse
		if err := json.Unmarshal(body, &scriptResponse); err != nil {
			return scripts.Script{}, fmt.Errorf("unmarshal erro: %s", err.Error())
		}

		return scriptResponse.ToEntity(), nil

	})

	if err != nil {
		return scripts.Script{}, err
	}

	return response.(scripts.Script), nil
}

func (a AccountScriptsApi) Close() {
	a.httpClient.CloseIdleConnections()
}
