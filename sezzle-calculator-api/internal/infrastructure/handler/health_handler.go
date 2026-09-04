package handler

import (
	"net/http"

	"github.com/labstack/echo/v5"
)

type HealthHandler struct{}

type HealthResponse struct {
	Status string `json:"status"`
}

const statusAlive = "ok"

func NewHealthHandler() *HealthHandler {
	return &HealthHandler{}
}

// LivenessHandler godoc
// @Summary      Liveness probe
// @Description  Reports that the process is up and serving. It performs no
// @Description  dependency checks, so a failure means the server itself is gone.
// @Tags         health
// @Produce      json
// @Success      200  {object} HealthResponse
// @Router       /health/live [get]
func (hh *HealthHandler) LivenessHandler(c *echo.Context) error {
	return c.JSON(http.StatusOK, HealthResponse{Status: statusAlive})
}

func (hh *HealthHandler) Register(e *echo.Echo) {
	e.GET("/health/live", hh.LivenessHandler)
}
