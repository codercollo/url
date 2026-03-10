package handler

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// HealthResponse represents the API health status response
type HealthResponse struct {
	Status string            `json:"status"`
	Checks map[string]string `json:"checks"`
}

func (h *Handler) Health(c *gin.Context) {
	checks := make(map[string]string)
	status := "ok"

	//Check Postgres
	ctx, cancel := context.WithTimeout(c.Request.Context(), 2*time.Second)
	defer cancel()

	if err := h.db.PingContext(ctx); err != nil {
		checks["postgres"] = "error" + err.Error()
		status = "degraded"
	} else {
		checks["postgres"] = "ok"
	}

	//Check Redis
	ctx2, cancel := context.WithTimeout(c.Request.Context(), 3*time.Second)
	defer cancel()

	if err := h.rdb.Ping(ctx2).Err(); err != nil {
		checks["redis"] = "error:" + err.Error()
		status = "degraded"
	} else {
		checks["redis"] = "ok"
	}

	code := http.StatusOK
	if status == "degraded" {
		code = http.StatusServiceUnavailable
	}

	c.JSON(code, HealthResponse{
		Status: status,
		Checks: checks,
	})
}
