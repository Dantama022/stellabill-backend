package handlers

import (
	"context"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
)

const (
	ServiceName        = "stellarbill-backend"
	StatusReady        = "ready"
	StatusDegraded     = "degraded"
	StatusUnavailable  = "unavailable"
	MaxRetries         = 3
	MaxDatabaseTimeout = 2 * time.Second
)

var InitialBackoff = 100 * time.Millisecond

// DBPinger abstracts database connectivity checks for readiness.
type DBPinger interface {
	PingContext(ctx context.Context) error
}

// HealthResponse is returned by readiness checks.
type HealthResponse struct {
	Status       string            `json:"status"`
	Service      string            `json:"service"`
	Timestamp    string            `json:"timestamp"`
	Dependencies map[string]string `json:"dependencies,omitempty"`
}

// Health exposes a basic liveness response.
func Health(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":  "ok",
		"service": ServiceName,
	})
}

// Health is kept as a method for existing handler tests.
func (h *Handler) Health(c *gin.Context) {
	Health(c)
}

// OutboxStats placeholder endpoint.
func OutboxStats(c *gin.Context) {
	c.JSON(http.StatusServiceUnavailable, gin.H{"error": "outbox manager not available"})
}

// PublishTestEvent placeholder endpoint.
func PublishTestEvent(c *gin.Context) {
	c.JSON(http.StatusServiceUnavailable, gin.H{"error": "outbox manager not available"})
}

// ReadinessHandler checks dependency readiness.
func ReadinessHandler(db DBPinger) gin.HandlerFunc {
	return func(c *gin.Context) {
		deps := map[string]string{"database": checkDatabase(db)}
		overallStatus := deriveOverallStatus(deps)

		resp := HealthResponse{
			Status:       overallStatus,
			Service:      ServiceName,
			Timestamp:    time.Now().UTC().Format(time.RFC3339),
			Dependencies: deps,
		}

		statusCode := http.StatusOK
		if overallStatus == StatusDegraded || overallStatus == StatusUnavailable {
			statusCode = http.StatusServiceUnavailable
		}

		c.JSON(statusCode, resp)
	}
}

func checkDatabase(db DBPinger) string {
	if os.Getenv("DATABASE_URL") == "" {
		return "not_configured"
	}
	if db == nil {
		return "down"
	}

	var lastErr error
	for i := 0; i < MaxRetries; i++ {
		ctx, cancel := context.WithTimeout(context.Background(), MaxDatabaseTimeout)
		lastErr = db.PingContext(ctx)
		cancel()
		if lastErr == nil {
			return "up"
		}
		if i < MaxRetries-1 {
			time.Sleep(InitialBackoff)
		}
	}

	if lastErr == context.DeadlineExceeded {
		return "timeout"
	}
	return "down"
}

func deriveOverallStatus(deps map[string]string) string {
	for _, status := range deps {
		if status == "down" || status == "timeout" {
			return StatusDegraded
		}
	}
	return StatusReady
}
