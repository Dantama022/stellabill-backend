package middleware

import (
	"fmt"
	"net/http"
	"os"
	"strings"

	"stellarbill-backend/internal/auth"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// ErrorEnvelope for auth errors
type ErrorEnvelope struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	TraceID string `json:"trace_id"`
}

// respondAuthError is a helper to respond with auth errors in the standard envelope format
func respondAuthError(c *gin.Context, message string) {
	c.Header("Content-Type", "application/json; charset=utf-8")
	
	traceID := c.GetString("traceID")
	if traceID == "" {
		traceID = uuid.New().String()
	}

	envelope := ErrorEnvelope{
		Code:    "UNAUTHORIZED",
		Message: message,
		TraceID: traceID,
	}

	c.AbortWithStatusJSON(http.StatusUnauthorized, envelope)
}

// AuthMiddleware validates the Authorization header (Bearer JWT) with hardened settings.
func AuthMiddleware(jwtSecret string) gin.HandlerFunc {
	// Initialize hardened config
	cfg := auth.Config{
		Secret:    []byte(jwtSecret),
		Issuer:    os.Getenv("JWT_ISSUER"),   // Should be configured
		Audience:  os.Getenv("JWT_AUDIENCE"), // Should be configured
		Algorithm: "HS256",                   // Explicit algorithm
	}

	// Use dev defaults if not provided (not for production)
	if cfg.Issuer == "" {
		cfg.Issuer = "stellabill"
	}
	if cfg.Audience == "" {
		cfg.Audience = "api-clients"
	}

	return auth.GinJWTMiddleware(cfg)
}
