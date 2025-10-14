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

		c.Locals("userID", claims.UserID)
		c.Locals("userName", claims.Name)

		return c.Next()
	}
}

// OptionalJWTMiddleware tenta validar JWT tokens se presente, mas não falha se ausente
// Útil para endpoints que podem funcionar com ou sem autenticação
func OptionalJWTMiddleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
		authHeader := c.Get("Authorization")

		// Se não há header de autorização, continua sem autenticação
		if authHeader == "" {
			return c.Next()
		}

		// Se há header mas não está no formato correto, continua sem autenticação
		if !strings.HasPrefix(authHeader, "Bearer ") {
			return c.Next()
		}

		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		if tokenString == "" {
			return c.Next()
		}

		// Tenta verificar o token
		claims, err := verifyJWT(tokenString)
		if err != nil {
			// Token inválido, mas não falha - apenas continua sem autenticação
			return c.Next()
		}

		// Token válido - adiciona informações do usuário ao contexto
		c.Locals("userID", claims.UserID)
		c.Locals("userName", claims.Name)

		return c.Next()
	}
}

// GetUserFromContext extrai informações do usuário do fiber context
func GetUserFromContext(c *fiber.Ctx) (userID int64, name string, ok bool) {
	userIDInterface := c.Locals("userID")
	nameInterface := c.Locals("userName")

	if userIDInterface == nil || nameInterface == nil {
		return 0, "", false
	}

	userID, userIDOk := userIDInterface.(int64)
	name, nameOk := nameInterface.(string)

	if !userIDOk || !nameOk {
		return 0, "", false
	}

	return userID, name, true
}
