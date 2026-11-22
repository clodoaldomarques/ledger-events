package ledger

import (
	"net/http"

	"github.com/clodoaldomarques/ledger-events/internal/adapters/inbound/http/shared"
	"github.com/clodoaldomarques/ledger-events/internal/core/ledger/ports"
	"github.com/labstack/echo/v4"
)

type Handler struct {
	uc ports.EventUseCase
}

func NewHandler(uc ports.EventUseCase) *Handler {
	return &Handler{uc: uc}
}

func (h Handler) CreateEvent(c echo.Context) error {
	orgID, cid := shared.GetHeaders(c)
	ctx := c.Request().Context()

	evt := new(EventRequest)
	if err := c.Bind(evt); err != nil {
		return echo.ErrBadRequest
	}

	if err := evt.Validate(); err != nil {
		return shared.MapError(c, err)
	}

	saved, err := h.uc.CreateEvent(ctx, cid, evt.ToEntity(orgID), evt.Amounts, evt.Fees)
	if err != nil {
		return shared.MapError(c, err)
	}

	return c.JSON(http.StatusCreated, buildEventResponse(saved))
}
