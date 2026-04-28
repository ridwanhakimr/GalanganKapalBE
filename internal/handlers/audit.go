package handlers

import (
	"github.com/gofiber/fiber/v2"
	"github.com/shipyard-system/backend/internal/config"
	"github.com/shipyard-system/backend/internal/models"
)

// Get Audit Logs (Admin only)
func GetAuditLogs(c *fiber.Ctx) error {
	var logs []models.AuditLog
	// Fetch with User relation and order by newest
	if err := config.DB.Preload("User").Order("created_at desc").Limit(100).Find(&logs).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Gagal mengambil audit log"})
	}
	return c.JSON(logs)
}
