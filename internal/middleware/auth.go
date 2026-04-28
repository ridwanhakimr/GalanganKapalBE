package middleware

import (
	"os"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/shipyard-system/backend/internal/models"
	"github.com/shipyard-system/backend/pkg/utils"
)

func Protected() fiber.Handler {
	return func(c *fiber.Ctx) error {
		authHeader := c.Get("Authorization")
		if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Missing or invalid token",
			})
		}

		tokenString := strings.Split(authHeader, " ")[1]
		secret := os.Getenv("JWT_SECRET")

		claims, err := utils.VerifyToken(tokenString, secret)
		if err != nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Unauthorized or token expired",
			})
		}

		// Save claims to context
		c.Locals("user_id", claims.UserID)
		c.Locals("role", claims.Role)

		return c.Next()
	}
}

func RequireRole(roles ...models.UserRole) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Read role from locals set by Protected middleware
		userRoleStr := c.Locals("role")
		if userRoleStr == nil {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"error": "Missing role in token",
			})
		}

		userRole := userRoleStr.(models.UserRole)

		roleAllowed := false
		for _, role := range roles {
			if role == userRole {
				roleAllowed = true
				break
			}
		}

		if !roleAllowed {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"error": "Forbidden: Requires different privileges",
			})
		}

		return c.Next()
	}
}
