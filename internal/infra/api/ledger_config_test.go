package api

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/clodoaldomarques/ledger-events/config"
	"github.com/clodoaldomarques/ledger-events/internal/domain/configs"
	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestAccountconfigsApi_FindScriptByLevel(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path := strings.Replace(r.URL.Path, "/v1/ledger/config/", "", 1)
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
			w.Write([]byte(`{"message": "ledger config not found"}`))
		}
	}))
	defer server.Close()

	type args struct {
		cid             string
		processing_code string
		orgID           string
		programID       int64
	}
	tests := []struct {
		name  string
		setup func(ctrl *gomock.Controller) *LedgerConfigApi
		args  func() args
		want  func(t *testing.T, sc configs.Config, e error)
	}{
		{
			name: "success - retrieve ledger config",
			setup: func(ctrl *gomock.Controller) *LedgerConfigApi {
				config.New(config.WithLedgerConfigApiUrl(server.URL))
				return New(context.Background())
			},
			args: func() args {
				return args{
					cid:             uuid.NewString(),
					processing_code: "b-612",
					orgID:           "TN-Test",
					programID:       007,
				}
			},
			want: func(t *testing.T, sc configs.Config, e error) {
				assert.Nil(t, e)
				assert.NotNil(t, sc)
				assert.Equal(t, "725a69c4-ec2c-4f6e-918c-d38c75e37b71", sc.ConfigID)
				assert.Equal(t, configs.TenantLevel, sc.Level)
			},
		},
		{
			name: "error - retrieve ledger config",
			setup: func(ctrl *gomock.Controller) *LedgerConfigApi {
				config.New(config.WithLedgerConfigApiUrl(server.URL))
				return New(context.Background())
			},
			args: func() args {
				return args{
					cid:             uuid.NewString(),
					processing_code: "b-615",
					orgID:           "TN-Test",
					programID:       007,
				}
			},
			want: func(t *testing.T, sc configs.Config, e error) {
				assert.NotNil(t, e)
				assert.Equal(t, "ledger config not found", e.Error())
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			api := tt.setup(ctrl)
			sc, err := api.FindConfigByLevel(context.Background(), tt.args().cid, tt.args().processing_code, tt.args().orgID, tt.args().programID)
			tt.want(t, sc, err)
		})
	}
}

func fakeScript(processing_code string) []byte {
	configs := map[string]string{
		"b-612": `{ "config_id": "725a69c4-ec2c-4f6e-918c-d38c75e37b71", "level": "tenant", "process_code": "b-612", "org_id": "TN-Test", "description": "Compra a vista", "scripts": [ { "script_id": 101, "flow": "regular", "description": "Compra a vista - Cartão", "expression": "Amount.amount + Fee.iof"}, {"script_id": 101, "flow": "migration", "description": "Compra a vista - PIX", "expression": "Amount.amount" }], "enable": true, "created_at": "2026-05-07T01:28:15.15791924Z", "updated_at": "2026-05-07T01:28:15.157922523Z", "version": 1 }`,
	}
	if sc, has := configs[processing_code]; has {
		return []byte(sc)
	}
	return nil
}
