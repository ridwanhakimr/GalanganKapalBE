package handlers

import (
	"encoding/json"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/shipyard-system/backend/internal/config"
	"github.com/shipyard-system/backend/internal/models"
	"gorm.io/gorm/clause"
)

type CreateRequestInput struct {
	ItemID   string `json:"item_id"`
	Quantity int    `json:"quantity"`
	Notes    string `json:"notes"`
}

// @desc Get all requests
// @route GET /api/v1/requests
func GetRequests(c *fiber.Ctx) error {
	var requests []models.ItemRequest
	if err := config.DB.Preload("Requester").Preload("Item").Order("requested_at desc").Find(&requests).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Gagal mengambil data permintaan"})
	}
	return c.JSON(requests)
}

// @desc Create a new item request (Staff)
// @route POST /api/v1/requests
func CreateRequest(c *fiber.Ctx) error {
	var input CreateRequestInput
	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid input"})
	}

	itemID, err := uuid.Parse(input.ItemID)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid item ID"})
	}

	userID := c.Locals("user_id").(uuid.UUID)

	// Start Database Transaction
	tx := config.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	var item models.Item
	// FOR UPDATE locks the row to prevent race conditions
	if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&item, "id = ?", itemID).Error; err != nil {
		tx.Rollback()
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Item not found"})
	}

	// Validation
	if item.Quantity < input.Quantity {
		tx.Rollback()
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Insufficient stock available"})
	}

	// Create Request
	req := models.ItemRequest{
		RequesterID:       userID,
		ItemID:            item.ID,
		QuantityRequested: input.Quantity,
		Notes:             input.Notes,
		Status:            models.RequestPending,
	}

	if err := tx.Create(&req).Error; err != nil {
		tx.Rollback()
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to create request"})
	}

	// Log Audit
	payload, _ := json.Marshal(req)
	audit := models.AuditLog{
		UserID:    &userID,
		RequestID: &req.ID,
		Action:    "REQUEST_CREATED",
		Payload:   string(payload),
		IPAddress: c.IP(),
	}
	tx.Create(&audit)

	// Commit Transaction
	tx.Commit()

	// In real world, trigger async notification here (e.g. Notify Supervisor)

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"message": "Request successfully created",
		"data":    req,
	})
}

// @desc Approve a request (Supervisor only)
// @route PATCH /api/v1/requests/:id/approve
func ApproveRequest(c *fiber.Ctx) error {
	reqID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request ID format"})
	}

	approverID := c.Locals("user_id").(uuid.UUID)

	tx := config.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	var req models.ItemRequest
	if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&req, "id = ?", reqID).Error; err != nil {
		tx.Rollback()
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Request not found"})
	}

	if req.Status != models.RequestPending {
		tx.Rollback()
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Only pending requests can be approved"})
	}

	// Lock item before deduction
	var item models.Item
	if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&item, "id = ?", req.ItemID).Error; err != nil {
		tx.Rollback()
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Item not found"})
	}

	// Check stock one more time!
	if item.Quantity < req.QuantityRequested {
		tx.Rollback()
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Insufficient stock to approve. Item may have been taken by another request."})
	}

	// Update Item stock
	item.Quantity = item.Quantity - req.QuantityRequested
	if err := tx.Save(&item).Error; err != nil {
		tx.Rollback()
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to update item stock"})
	}

	// Update Request status
	now := time.Now()
	req.Status = models.RequestApproved
	req.ApproverID = &approverID
	req.ApprovedAt = &now

	if err := tx.Save(&req).Error; err != nil {
		tx.Rollback()
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to update request status"})
	}

	// Insert Stock Movement
	movement := models.StockMovement{
		ItemID:       item.ID,
		RequestID:    &req.ID,
		QtyChange:    -req.QuantityRequested,
		MovementType: models.MovementOut,
	}
	if err := tx.Create(&movement).Error; err != nil {
		tx.Rollback()
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to record stock movement"})
	}

	// Log Audit
	payload, _ := json.Marshal(map[string]interface{}{
		"request_id": req.ID,
		"item_id":    item.ID,
		"deducted":    req.QuantityRequested,
		"new_stock":  item.Quantity,
	})
	audit := models.AuditLog{
		UserID:    &approverID,
		RequestID: &req.ID,
		Action:    "REQUEST_APPROVED",
		Payload:   string(payload),
		IPAddress: c.IP(),
	}
	tx.Create(&audit)

	tx.Commit()

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "Request approved successfully",
		"request": req,
		"item_available_qty": item.Quantity,
	})
}

// @desc Reject a request (Supervisor only)
// @route PATCH /api/v1/requests/:id/reject
func RejectRequest(c *fiber.Ctx) error {
	reqID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request ID format"})
	}

	approverID := c.Locals("user_id").(uuid.UUID)

	tx := config.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	var req models.ItemRequest
	if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&req, "id = ?", reqID).Error; err != nil {
		tx.Rollback()
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Request not found"})
	}

	if req.Status != models.RequestPending {
		tx.Rollback()
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Only pending requests can be rejected"})
	}

	// Update Request status to Rejected
	now := time.Now()
	req.Status = models.RequestRejected
	req.ApproverID = &approverID
	req.ApprovedAt = &now // Using same timestamp field for decision time

	if err := tx.Save(&req).Error; err != nil {
		tx.Rollback()
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to update request status"})
	}

	// Log Audit
	payload, _ := json.Marshal(map[string]interface{}{
		"request_id": req.ID,
		"item_id":    req.ItemID,
		"status":     "REJECTED",
	})
	audit := models.AuditLog{
		UserID:    &approverID,
		RequestID: &req.ID,
		Action:    "REQUEST_REJECTED",
		Payload:   string(payload),
		IPAddress: c.IP(),
	}
	tx.Create(&audit)

	tx.Commit()

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "Request rejected successfully",
		"request": req,
	})
}
