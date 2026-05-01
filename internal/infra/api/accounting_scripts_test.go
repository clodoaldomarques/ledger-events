package api

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/clodoaldomarques/ledger-events/configs"
	"github.com/clodoaldomarques/ledger-events/internal/domain/scripts"
	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestAccountScriptsApi_FindScriptByLevel(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path := strings.Replace(r.URL.Path, "/v1/accounting/scripts/", "", 1)
		filters := strings.Split(path, "/")

		assert.Equal(t, "application/json", r.Header.Get("Content-Type"))
		assert.NotEmpty(t, r.Header.Get("x-cid"))
		assert.NotEmpty(t, r.Header.Get("x-tenant"))

		w.Header().Set("Content-Type", "application/json")

		sc := fakeScript(filters[0])
		if sc != nil {
			w.WriteHeader(http.StatusOK)
			w.Write(fakeScript(filters[0]))
		} else {
			w.WriteHeader(http.StatusNotFound)
			w.Write([]byte(`{"message": "accounting script not found"}`))
		}
	}))
	defer server.Close()

	type args struct {
		cid         string
		eventTypeID string
		orgID       string
		programID   int64
	}
	tests := []struct {
		name  string
		setup func(ctrl *gomock.Controller) *AccountScriptsApi
		args  func() args
		want  func(t *testing.T, sc scripts.Script, e error)
	}{
		{
			name: "success - retrieve accounting script",
			setup: func(ctrl *gomock.Controller) *AccountScriptsApi {
				configs.New(configs.WithAccountScriptsApiUrl(server.URL))
				return New(context.Background())
			},
			args: func() args {
				return args{
					cid:         uuid.NewString(),
					eventTypeID: "b-612",
					orgID:       "TN-Test",
					programID:   007,
				}
			},
			want: func(t *testing.T, sc scripts.Script, e error) {
				assert.Nil(t, e)
				assert.NotNil(t, sc)
				assert.Equal(t, "pismo-B-612", sc.ScriptID)
				assert.Equal(t, scripts.PlatformLevel, sc.Level)
			},
		},
		{
			name: "error - retrieve accounting script",
			setup: func(ctrl *gomock.Controller) *AccountScriptsApi {
				configs.New(configs.WithAccountScriptsApiUrl(server.URL))
				return New(context.Background())
			},
			args: func() args {
				return args{
					cid:         uuid.NewString(),
					eventTypeID: "b-615",
					orgID:       "TN-Test",
					programID:   007,
				}
			},
			want: func(t *testing.T, sc scripts.Script, e error) {
				assert.NotNil(t, e)
				assert.Equal(t, "accounting script not found", e.Error())
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			api := tt.setup(ctrl)
			sc, err := api.FindScriptByLevel(context.Background(), tt.args().cid, tt.args().eventTypeID, tt.args().orgID, tt.args().programID)
			tt.want(t, sc, err)
		})
	}
}

func fakeScript(eventTypeID string) []byte {
	scripts := map[string]string{
		"b-612": `{"script_id": "pismo-B-612", "level": "platform", "event_type_id": "B-612", "org_id": "PISMO", "description": "Parcelamento xpto", "entries": [{"entry_type_id": 1001, "flow": "regular", "description": "Parcelamento", "amount_name": "amount_name",	"expression": "Amount.amount"}, {"entry_type_id": 1002,	"flow": "migration", "description": "Parcelamento",	"amount_name": "duo_date", "expression": "Amount.amount + Fee.iof", "parameter": {"name": "classificatio","value": "2"}}], "enable": true,	"created_at": "2025-08-16T12:20:13Z", "updated_at": "2025-08-16T12:20:13Z",	"version": 2}`,
	}
	if sc, has := scripts[eventTypeID]; has {
		return []byte(sc)
	}
	return nil
}
