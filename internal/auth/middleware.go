package auth

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

const RoleContextKey = "role"

// Extract role from JWT claims in request context
// Previously relied on header-based role extraction; now uses validated JWT claims
func ExtractRole(c *gin.Context) Role {
	// Only get from context (set by hardened JWT middleware)
	if role, exists := c.Get(RoleContextKey); exists {
		if roleStr, ok := role.(string); ok {
			return Role(roleStr)
		}
	}

	return ""
}

// RequirePermission middleware enforces role-based access control
// Validates that the authenticated user has the required permission
func RequirePermission(permission Permission) gin.HandlerFunc {
	return func(c *gin.Context) {
		role := ExtractRole(c)

		if role == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "missing role - ensure JWT middleware is applied",
			})
			return
		}

		if !HasPermission(role, permission) {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
				"error": "insufficient permissions for this operation",
			})
			return
		}

		c.Set(RoleContextKey, role)
		c.Next()
	}
}
