package main

import (
	"log"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/shipyard-system/backend/internal/config"
	"github.com/shipyard-system/backend/internal/handlers"
	"github.com/shipyard-system/backend/internal/middleware"
	"github.com/shipyard-system/backend/internal/models"
)

func main() {
	// Initialize configurations (env)
	appConfig := config.LoadConfig()

	// Initialize Database
	config.ConnectDatabase(appConfig.DatabaseURL)
	
	// Check if this is a migration run
	// For production, maybe it's better to put this under CLI command instead of auto-running on boot
	if config.DB != nil {
		config.MigrateDB()
	}

	// Initialize Fiber App
	app := fiber.New(fiber.Config{
		AppName: "Shipyard Management API",
	})

	// Middlewares
	app.Use(logger.New())
	app.Use(cors.New())

	// Serve Static Files for Image Uploads
	app.Static("/uploads", "./uploads")

	// Basic route
	app.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"status":  "ok",
			"message": "Shipyard System API is running",
		})
	})

	// Setup Routes
	api := app.Group("/api/v1")
	api.Post("/auth/login", handlers.Login)
	api.Get("/setup", handlers.SetupSeedUsers) // Endpoint Khusus cetak Akun

	// Proteksi Route (hanya yang sudah login)
	protected := api.Group("/", middleware.Protected())

	// Endpoint Inventory
	protected.Get("/warehouses", handlers.GetWarehouses)
	protected.Get("/categories", handlers.GetCategories)
	protected.Get("/items", handlers.GetItems)
	
	// Endpoint Khusus Admin (RBAC)
	adminOnly := protected.Group("", middleware.RequireRole(models.RoleAdmin))
	
	// Master Data: Warehouses
	adminOnly.Post("/warehouses", handlers.CreateWarehouse)
	adminOnly.Put("/warehouses/:id", handlers.UpdateWarehouse)
	adminOnly.Delete("/warehouses/:id", handlers.DeleteWarehouse)
	
	// Master Data: Categories
	adminOnly.Post("/categories", handlers.CreateCategory)
	adminOnly.Put("/categories/:id", handlers.UpdateCategory)
	adminOnly.Delete("/categories/:id", handlers.DeleteCategory)
	
	adminOnly.Post("/items", handlers.CreateItem)
	adminOnly.Get("/audit", handlers.GetAuditLogs)
	adminOnly.Get("/stock-movements", handlers.GetStockMovements)

	// Endpoint Requests
	requests := protected.Group("/requests")
	requests.Get("/", handlers.GetRequests)
	requests.Post("/", handlers.CreateRequest)
	// SPV & Admin Approval & Rejection
	requests.Patch("/:id/approve", middleware.RequireRole(models.RoleSupervisor), handlers.ApproveRequest)
	requests.Patch("/:id/reject", middleware.RequireRole(models.RoleSupervisor), handlers.RejectRequest)

	// Start server
	log.Printf("Server starting on port %s", appConfig.AppPort)
	log.Fatal(app.Listen(":" + appConfig.AppPort))
}

