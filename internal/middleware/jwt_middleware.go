package middleware

import (
	"exam/internal/repository"
	"exam/internal/utils"
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo/v4"
)

// JWTAuthMiddleware checks for a valid JWT token in the Authorization header.
func JWTAuthMiddleware(deviceRepo repository.DeviceRepository) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			authHeader := c.Request().Header.Get("Authorization")
			if authHeader == "" {
				return utils.ErrorResponse(c, http.StatusUnauthorized, "Authorization header is missing")
			}

			headerParts := strings.Split(authHeader, " ")
			if len(headerParts) != 2 || strings.ToLower(headerParts[0]) != "bearer" {
				return utils.ErrorResponse(c, http.StatusUnauthorized, "Authorization header format must be Bearer {token}")
			}

			tokenString := headerParts[1]

			jwtSecret := os.Getenv("JWT_SECRET")
			if jwtSecret == "" {
				jwtSecret = "supersecretjwtkey" // Should match the one in auth_service.go
			}

			token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
				// Don't forget to validate the alg is what you expect: `jwt.SigningMethodHS256`
				if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
					return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
				}
				return []byte(jwtSecret), nil
			})

			if err != nil {
				return utils.ErrorResponse(c, http.StatusUnauthorized, "Invalid or expired token: "+err.Error())
			}

			if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
				// Check if JTI is valid in the devices table
				jti, jtiOk := claims["jti"].(string)
				if !jtiOk || jti == "" {
					return utils.ErrorResponse(c, http.StatusUnauthorized, "JTI claim missing or invalid")
				}

				device, err := deviceRepo.GetDeviceByJTI(jti) // Use GetDeviceByJTI
				if err != nil {
					return utils.ErrorResponse(c, http.StatusInternalServerError, "Database error during JTI validation")
				}
				if device == nil {
					return utils.ErrorResponse(c, http.StatusUnauthorized, "Invalid JTI: Device not found")
				}

				if device.LogoutAt != nil {
					return utils.ErrorResponse(c, http.StatusUnauthorized, "Token has been logged out")
				}

				// Set user information in context
				c.Set("userID", claims["id"])
				c.Set("userUUID", claims["uuid"])
				c.Set("userEmail", claims["email"])
				c.Set("userRole", claims["role"])
				c.Set("jti", jti)
				c.Set("exp", claims["exp"])

				return next(c)
			}

			return utils.ErrorResponse(c, http.StatusUnauthorized, "Invalid token claims")
		}
	}
}
