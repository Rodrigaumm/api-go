package handlers

import (
	"strings"

	"github.com/gofiber/fiber/v2"
)

// JWTMiddleware valida JWT tokens e adiciona user info no context
func JWTMiddleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
		authHeader := c.Get("Authorization")
		if authHeader == "" {
			return c.Status(401).JSON(fiber.Map{
				"error": "Authorization header required",
			})
		}

		if !strings.HasPrefix(authHeader, "Bearer ") {
			return c.Status(401).JSON(fiber.Map{
				"error": "Invalid authorization header format",
			})
		}

		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		if tokenString == "" {
			return c.Status(401).JSON(fiber.Map{
				"error": "Token required",
			})
		}

		claims, err := verifyJWT(tokenString)
		if err != nil {
			return c.Status(401).JSON(fiber.Map{
				"error": "Invalid or expired token",
			})
		}

		c.Locals("user_id", claims.UserID)
		c.Locals("user_email", claims.Email)

		return c.Next()
	}
}

// GetUserFromContext extrai informações do usuário do fiber context
func GetUserFromContext(c *fiber.Ctx) (userID int64, email string, ok bool) {
	userIDInterface := c.Locals("user_id")
	emailInterface := c.Locals("user_email")

	if userIDInterface == nil || emailInterface == nil {
		return 0, "", false
	}

	userID, userIDOk := userIDInterface.(int64)
	email, emailOk := emailInterface.(string)

	if !userIDOk || !emailOk {
		return 0, "", false
	}

	return userID, email, true
}
