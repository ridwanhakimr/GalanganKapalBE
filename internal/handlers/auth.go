package handlers

import (
	"os"

	"github.com/gofiber/fiber/v2"
	"github.com/shipyard-system/backend/internal/config"
	"github.com/shipyard-system/backend/internal/models"
	"github.com/shipyard-system/backend/pkg/utils"
)

type LoginInput struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func Login(c *fiber.Ctx) error {
	var input LoginInput
	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid form input"})
	}

	var user models.User
	if err := config.DB.Where("email = ?", input.Email).First(&user).Error; err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Email atau password salah"})
	}

	if !utils.CheckPasswordHash(input.Password, user.PasswordHash) {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Email atau password salah"})
	}

	secret := os.Getenv("JWT_SECRET")
	token, err := utils.GenerateToken(&user, secret)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Gagal menghasilkan token session"})
	}

	return c.JSON(fiber.Map{
		"token": token,
		"user": fiber.Map{
			"id":    user.ID,
			"name":  user.Name,
			"role":  user.Role,
			"email": user.Email,
		},
	})
}

// Endpoint ini sengaja dibuat agar Anda bisa dengan mudah membuat data sample/login untuk tahap ini
// Karena password harus tersimpan dalam bentuk Bcrypt hash (disandikan)
func SetupSeedUsers(c *fiber.Ctx) error {
	var count int64
	config.DB.Model(&models.User{}).Count(&count)
	
	if count > 0 {
		return c.Status(400).JSON(fiber.Map{"message": "Database telah terisi, fitur seed dinonaktifkan."})
	}

	hash, _ := utils.HashPassword("admin123") // Kata sandi seragam untuk 3 akun
	
	admin := models.User{Name: "Bapak Admin", Email: "admin@ship.com", PasswordHash: hash, Role: models.RoleAdmin, IsActive: true}
	supervisor := models.User{Name: "Pak Spv", Email: "spv@ship.com", PasswordHash: hash, Role: models.RoleSupervisor, IsActive: true}
	staff := models.User{Name: "Mas Staff", Email: "staff@ship.com", PasswordHash: hash, Role: models.RoleStaff, IsActive: true}

	config.DB.Create(&admin)
	config.DB.Create(&supervisor)
	config.DB.Create(&staff)

	return c.JSON(fiber.Map{
		"message": "Berhasil memompa (seed) data percobaan!",
		"users_created": []string{
			"admin@ship.com (Ketik role: Admin)",
			"spv@ship.com (Ketik role: Supervisor)",
			"staff@ship.com (Ketik role: Staff)",
		},
		"default_password": "admin123",
	})
}
