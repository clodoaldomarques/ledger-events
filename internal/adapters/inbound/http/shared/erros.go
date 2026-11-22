package shared

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

type ErrResponse struct {
	Message string `json:"message"`
}

func (e ErrResponse) Error() string {
	return e.Message
}

// MapError mantém o comportamento atual: qualquer falha do caso de uso vira 400.
func MapError(c echo.Context, err error) error {
	return c.JSON(http.StatusBadRequest, ErrResponse{Message: err.Error()})
}
