package middleware

import (
	"exam/internal/utils"
	"net/http"

	"github.com/casbin/casbin/v2"
	"github.com/labstack/echo/v4"
)

// CasbinAuthMiddleware checks if the user has permission to access the requested resource.
func CasbinAuthMiddleware(enforcer *casbin.Enforcer) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			// Get user role from context (set by JWTAuthMiddleware)
			role, ok := c.Get("userRole").(string)
			if !ok || role == "" {
				return utils.ErrorResponse(c, http.StatusForbidden, "User role not found in context")
			}

			// Get requested resource and action
			obj := c.Request().URL.Path
			act := c.Request().Method

			// Check if the user has permission
			ok, err := enforcer.Enforce(role, obj, act)
			if err != nil {
				return utils.ErrorResponse(c, http.StatusInternalServerError, "Authorization error")
			}

			if !ok {
				return utils.ErrorResponse(c, http.StatusForbidden, "Forbidden: You do not have permission to access this resource")
			}

			return next(c)
		}
	}
}
