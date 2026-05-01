package accounting

import (
	"net/http"

	"github.com/clodoaldomarques/ledger-events/internal/domain/events"
	"github.com/clodoaldomarques/ledger-events/internal/infra/api"
	"github.com/clodoaldomarques/ledger-events/internal/infra/db/mysql"
	"github.com/clodoaldomarques/ledger-events/internal/infra/message"
	"github.com/clodoaldomarques/ledger-events/internal/infra/rest/shared"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

func CreateEvent(c echo.Context) error {
	orgID, cid := getHeaders(c)
	ctx := c.Request().Context()

	a := api.New(ctx)
	defer a.Close()

	r := mysql.NewRepository(ctx)
	defer r.Close()

	t := message.New(ctx)
	defer t.Close()

	s := events.New(a, r, t)

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
