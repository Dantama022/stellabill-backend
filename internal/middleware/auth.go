package middleware

import (
	"net/http"
	"os"
	"strings"

	"stellarbill-backend/internal/auth"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"stellabill-backend/internal/auth"
)

type ErrorEnvelope struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	TraceID string `json:"trace_id"`
}

func respondAuthError(c *gin.Context, message string) {
	c.Header("Content-Type", "application/json; charset=utf-8")
<<<<<<< HEAD

=======
>>>>>>> upstream/main
	traceID := c.GetString("traceID")
	if traceID == "" {
		traceID = uuid.New().String()
	}

<<<<<<< HEAD
	c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
		"error":    message,
		"code":     "UNAUTHORIZED",
		"trace_id": traceID,
	})
}

// AuthMiddleware validates the Authorization header (Bearer JWT).
// On success it sets "callerID" in the Gin context and calls c.Next().
// On failure it aborts with 401 and a JSON error body.
func AuthMiddleware(jwtSecret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			respondAuthError(c, "missing authorization header")
			return
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
			respondAuthError(c, "invalid authorization header format")
			return
		}

		tokenStr := parts[1]
		var claims auth.Claims
		token, err := jwt.ParseWithClaims(tokenStr, &claims, func(t *jwt.Token) (interface{}, error) {
			if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, jwt.ErrSignatureInvalid
			}
			return []byte(jwtSecret), nil
		}, jwt.WithValidMethods([]string{"HS256", "HS384", "HS512"}))

		if err != nil || !token.Valid {
			respondAuthError(c, "invalid or expired token")
			return
		}

		// User identifier
		sub := claims.Subject
		if sub == "" {
			sub = claims.UserID
		}
		if sub == "" {
			respondAuthError(c, "token missing user identifier")
			return
		}

		// Tenant ID enforcement.
		tenantHeader := strings.TrimSpace(c.GetHeader("X-Tenant-ID"))
		tenantClaim := strings.TrimSpace(claims.Tenant)

		var tenantID string
		if tenantHeader != "" && tenantClaim != "" {
			if tenantHeader != tenantClaim {
				respondAuthError(c, "tenant mismatch")
				return
			}
			tenantID = tenantHeader
		} else if tenantHeader != "" {
			tenantID = tenantHeader
		} else if tenantClaim != "" {
			tenantID = tenantClaim
		} else {
			// If role is present and not admin, tenant is required.
			// If role is missing, we let it pass to let permission guards return 403.
			if claims.Role != "" && claims.Role != string(auth.RoleAdmin) {
				respondAuthError(c, "tenant id required")
				return
			}
			if claims.Role == string(auth.RoleAdmin) {
				tenantID = "system"
			}
		}

		c.Set("callerID", sub)
		if claims.Role != "" {
			c.Set("role", claims.Role)
		}

		if tenantID != "" {
			c.Set("tenantID", tenantID)
		}
		c.Next()
=======
	envelope := ErrorEnvelope{
		Code:    "UNAUTHORIZED",
		Message: message,
		TraceID: traceID,
	}
	c.AbortWithStatusJSON(http.StatusUnauthorized, envelope)
}

// AuthMiddleware creates a Gin middleware for JWT authentication with hardened settings.
// It supports both JWKS (asynchronous key rotation) and static secrets (HS256).
func AuthMiddleware(jwksCache *auth.JWKSCache, staticSecret string) gin.HandlerFunc {
	// Initialize hardened config
	cfg := auth.Config{
		Secret:    []byte(staticSecret),
		Issuer:    os.Getenv("JWT_ISSUER"),   // Should be configured
		Audience:  os.Getenv("JWT_AUDIENCE"), // Should be configured
		Algorithm: "HS256",                   // Explicit algorithm
		JWKS:      jwksCache,
>>>>>>> upstream/main
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

