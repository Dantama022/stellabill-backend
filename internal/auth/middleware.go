package auth

import (
	"net/http"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

const RoleContextKey = "role"
const RolesContextKey = "roles"

// ExtractRole returns the first available role from the request context
func ExtractRole(c *gin.Context) Role {
<<<<<<< HEAD
	if v, ok := c.Get(RoleContextKey); ok {
		if r, ok := v.(Role); ok {
			return r
		}
		if s, ok := v.(string); ok {
			return Role(s)
		}
	}
	role := c.GetHeader("X-Role")
	if role == "" {
=======
	roles := ExtractRoles(c)
	if len(roles) == 0 {
>>>>>>> upstream/main
		return ""
	}
	return roles[0]
}

// ExtractRoles returns all roles found in the request context (set by JWT middleware)
func ExtractRoles(c *gin.Context) []Role {
	// Only get from context (set by hardened JWT middleware)
	if roles := rolesFromContext(c); len(roles) > 0 {
		return roles
	}

	return nil
}

func rolesFromContext(c *gin.Context) []Role {
	if value, ok := c.Get(RolesContextKey); ok {
		switch typed := value.(type) {
		case []Role:
			return normalizeRoles(typed)
		case []string:
			roles := make([]Role, 0, len(typed))
			for _, role := range typed {
				roles = append(roles, Role(strings.TrimSpace(role)))
			}
			return normalizeRoles(roles)
		case string:
			return normalizeRoles([]Role{Role(strings.TrimSpace(typed))})
		}
	}

	if value, ok := c.Get(RoleContextKey); ok {
		switch typed := value.(type) {
		case Role:
			return normalizeRoles([]Role{typed})
		case string:
			return normalizeRoles([]Role{Role(strings.TrimSpace(typed))})
		}
	}

	return nil
}

func normalizeRoles(roles []Role) []Role {
	result := make([]Role, 0, len(roles))
	seen := map[Role]struct{}{}
	for _, role := range roles {
		role = Role(strings.TrimSpace(string(role)))
		if role == "" {
			continue
		}
		if _, ok := seen[role]; ok {
			continue
		}
		seen[role] = struct{}{}
		result = append(result, role)
	}
	return result
}
}

// RequirePermission middleware enforces role-based access control
// Validates that the authenticated user has the required permission
func RequirePermission(permission Permission) gin.HandlerFunc {
	return func(c *gin.Context) {
<<<<<<< HEAD
		role := ExtractRole(c)

		if role == "" || !HasPermission(role, permission) {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
				"error": "insufficient permissions",
=======
		roles := ExtractRoles(c)
		if len(roles) == 0 {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "missing role - ensure JWT middleware is applied",
			})
			return
		}

		for _, role := range roles {
			if HasPermission(role, permission) {
				c.Set(RoleContextKey, role)
				c.Set(RolesContextKey, roles)
				c.Next()
				return
			}
		}

		if len(roles) > 0 {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
				"error": "insufficient permissions for this operation",
>>>>>>> upstream/main
			})
			return
		}
	}
}

