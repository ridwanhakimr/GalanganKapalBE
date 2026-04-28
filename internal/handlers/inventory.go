package handlers

import (
	"fmt"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/shipyard-system/backend/internal/config"
	"github.com/shipyard-system/backend/internal/models"
)

// List Warehouses
func GetWarehouses(c *fiber.Ctx) error {
	var warehouses []models.Warehouse
	if err := config.DB.Find(&warehouses).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Gagal mengambil data gudang"})
	}
	return c.JSON(warehouses)
}

// List Categories
func GetCategories(c *fiber.Ctx) error {
	var categories []models.Category
	if err := config.DB.Find(&categories).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Gagal mengambil data kategori"})
	}
	return c.JSON(categories)
}

// List Items with pagination and relations
func GetItems(c *fiber.Ctx) error {
	var items []models.Item
	if err := config.DB.Preload("Warehouse").Preload("Category").Find(&items).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Gagal mengambil data inventaris"})
	}
	return c.JSON(items)
}

// Create Item with Image Upload
func CreateItem(c *fiber.Ctx) error {
	name := c.FormValue("name")
	sku := c.FormValue("sku")
	warehouseIDStr := c.FormValue("warehouse_id")
	categoryIDStr := c.FormValue("category_id")
	
	quantity, _ := strconv.Atoi(c.FormValue("quantity"))
	minStock, _ := strconv.Atoi(c.FormValue("min_stock", "5"))

	if name == "" || sku == "" || warehouseIDStr == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Nama, SKU, dan Gudang tidak boleh kosong"})
	}

	whID, err := uuid.Parse(warehouseIDStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Warehouse ID tidak valid"})
	}

	var catID *uuid.UUID
	if categoryIDStr != "" {
		parsedCat, err := uuid.Parse(categoryIDStr)
		if err == nil {
			catID = &parsedCat
		}
	}

	// Handle Image Upload
	var imageURL string
	file, err := c.FormFile("image")
	if err == nil && file != nil {
		// Validasi tipe extension
		ext := strings.ToLower(filepath.Ext(file.Filename))
		if ext == ".png" || ext == ".jpg" || ext == ".jpeg" || ext == ".webp" {
			// Generate unique filename
			fileName := fmt.Sprintf("%d-%s", time.Now().Unix(), uuid.New().String()+ext)
			filePath := fmt.Sprintf("./uploads/items/%s", fileName)

			if err := c.SaveFile(file, filePath); err != nil {
				return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Gagal menyimpan gambar"})
			}
			imageURL = "/uploads/items/" + fileName
		} else {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Format gambar tidak didukung"})
		}
	}

	item := models.Item{
		Name:        name,
		SKU:         sku,
		WarehouseID: whID,
		CategoryID:  catID,
		Quantity:    quantity,
		MinStock:    minStock,
		Status:      models.StatusTersedia,
		ImageURL:    imageURL,
	}

	if err := config.DB.Create(&item).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Gagal menyimpan item ke database (SKU mungkin duplikat)"})
	}

	// Reload with association
	config.DB.Preload("Warehouse").Preload("Category").First(&item, "id = ?", item.ID)

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"message": "Barang berhasil ditambahkan",
		"data":    item,
	})
}
