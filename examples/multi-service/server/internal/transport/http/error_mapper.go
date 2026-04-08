package http

import (
	"errors"
	"net/http"

	"github.com/HorseArcher567/octopus/examples/multi-service/server/internal/domain"
	"github.com/gin-gonic/gin"
)

type errorResponse struct {
	Error string `json:"error"`
}

func writeError(c *gin.Context, err error, notFoundMsg string) {
	if err == nil {
		return
	}
	switch {
	case errors.Is(err, domain.ErrNotFound):
		c.JSON(http.StatusNotFound, errorResponse{Error: notFoundMsg})
	case errors.Is(err, domain.ErrInvalidArgument):
		c.JSON(http.StatusBadRequest, errorResponse{Error: err.Error()})
	default:
		c.JSON(http.StatusInternalServerError, errorResponse{Error: err.Error()})
	}
}
