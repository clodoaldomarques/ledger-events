package shared

import (
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

// GetHeaders extrai x-tenant e x-cid, gerando um correlation id quando ausente.
func GetHeaders(c echo.Context) (string, string) {
	orgID := c.Request().Header.Get("x-tenant")
	cid := c.Request().Header.Get("x-cid")

	if cid == "" {
		cid = uuid.NewString()
	}
	return orgID, cid
}
