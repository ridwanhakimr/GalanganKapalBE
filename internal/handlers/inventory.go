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

// Create Warehouse (Admin)
func CreateWarehouse(c *fiber.Ctx) error {
	var input struct {
		Name        string `json:"name"`
		Location    string `json:"location"`
		Description string `json:"description"`
	}
	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid form input"})
	}

	if input.Name == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Nama gudang wajib diisi"})
	}

	warehouse := models.Warehouse{
		Name:        input.Name,
		Location:    input.Location,
		Description: input.Description,
		IsActive:    true,
	}

	if err := config.DB.Create(&warehouse).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Gagal menyimpan gudang"})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"message": "Gudang berhasil ditambahkan",
		"data":    warehouse,
	})
}

// Update Warehouse (Admin)
func UpdateWarehouse(c *fiber.Ctx) error {
	whID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid ID format"})
	}

	var warehouse models.Warehouse
	if err := config.DB.First(&warehouse, "id = ?", whID).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Warehouse not found"})
	}

	var input struct {
		Name        string `json:"name"`
		Location    string `json:"location"`
		Description string `json:"description"`
	}
	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid form input"})
	}

	if input.Name != "" {
		warehouse.Name = input.Name
	}
	warehouse.Location = input.Location
	warehouse.Description = input.Description

	config.DB.Save(&warehouse)
	return c.JSON(fiber.Map{"message": "Warehouse updated", "data": warehouse})
}

// Delete Warehouse (Admin)
func DeleteWarehouse(c *fiber.Ctx) error {
	whID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid ID format"})
	}

	// Cek apakah ada barang di gudang ini
	var count int64
	config.DB.Model(&models.Item{}).Where("warehouse_id = ?", whID).Count(&count)
	if count > 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Gudang masih berisi barang, tidak dapat dihapus"})
	}

	if err := config.DB.Delete(&models.Warehouse{}, "id = ?", whID).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to delete warehouse"})
	}
	return c.JSON(fiber.Map{"message": "Warehouse deleted"})
}

// List Categories
func GetCategories(c *fiber.Ctx) error {
	var categories []models.Category
	if err := config.DB.Find(&categories).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Gagal mengambil data kategori"})
	}
	return c.JSON(categories)
}

// Create Category (Admin)
func CreateCategory(c *fiber.Ctx) error {
	var input struct {
		Name        string `json:"name"`
		Description string `json:"description"`
	}
	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid form input"})
	}

	if input.Name == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Nama kategori wajib diisi"})
	}

	category := models.Category{
		Name:        input.Name,
		Description: input.Description,
	}

	if err := config.DB.Create(&category).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Gagal menyimpan kategori"})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"message": "Kategori berhasil ditambahkan",
		"data":    category,
	})
}

// Update Category (Admin)
func UpdateCategory(c *fiber.Ctx) error {
	catID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid ID format"})
	}

	var category models.Category
	if err := config.DB.First(&category, "id = ?", catID).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Category not found"})
	}

	var input struct {
		Name        string `json:"name"`
		Description string `json:"description"`
	}
	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid form input"})
	}

	if input.Name != "" {
		category.Name = input.Name
	}
	category.Description = input.Description

	config.DB.Save(&category)
	return c.JSON(fiber.Map{"message": "Category updated", "data": category})
}

// Delete Category (Admin)
func DeleteCategory(c *fiber.Ctx) error {
	catID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid ID format"})
	}

	var count int64
	config.DB.Model(&models.Item{}).Where("category_id = ?", catID).Count(&count)
	if count > 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Kategori masih digunakan pada beberapa barang, tidak dapat dihapus"})
	}

	if err := config.DB.Delete(&models.Category{}, "id = ?", catID).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to delete category"})
	}
	return c.JSON(fiber.Map{"message": "Category deleted"})
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

// Get Stock Movements (Ledger)
func GetStockMovements(c *fiber.Ctx) error {
	var movements []models.StockMovement
	if err := config.DB.Preload("Item").Order("moved_at desc").Find(&movements).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Gagal mengambil data buku besar"})
	}
	return c.JSON(movements)
}
