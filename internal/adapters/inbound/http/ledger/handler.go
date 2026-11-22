package ledger

import (
	"net/http"

	"github.com/clodoaldomarques/ledger-events/internal/adapters/inbound/http/shared"
	configApi "github.com/clodoaldomarques/ledger-events/internal/adapters/outbound/config"
	"github.com/clodoaldomarques/ledger-events/internal/adapters/outbound/mysql"
	"github.com/clodoaldomarques/ledger-events/internal/adapters/outbound/sns"
	"github.com/clodoaldomarques/ledger-events/internal/core/ledger/application"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

func CreateEvent(c echo.Context) error {
	orgID, cid := getHeaders(c)
	ctx := c.Request().Context()

	a := configApi.New(ctx)
	defer a.Close()

	r := mysql.NewRepository(ctx)
	defer r.Close()

	t := sns.New(ctx)
	defer t.Close(ctx)

	s := application.New(a, r, t)

	evt := new(EventRequest)
	if err := c.Bind(evt); err != nil {
		return echo.ErrBadRequest
	}

	if err := evt.Validate(); err != nil {
		return c.JSON(http.StatusBadRequest, shared.ErrResponse{Message: err.Error()})
	}

	saved, err := s.CreateEvent(ctx, cid, evt.ToEntity(orgID), evt.Amounts, evt.Fees)
	if err != nil {
		return c.JSON(http.StatusBadRequest, shared.ErrResponse{Message: err.Error()})
	}

	resp := buildEventResponse(saved)

	return c.JSON(http.StatusCreated, resp)
}

func getHeaders(c echo.Context) (string, string) {
	orgID := c.Request().Header.Get("x-tenant")
	cid := c.Request().Header.Get("x-cid")

	if cid == "" {
		cid = uuid.NewString()
	}
	return orgID, cid
}
