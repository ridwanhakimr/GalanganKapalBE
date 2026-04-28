package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type RequestStatus string

const (
	RequestPending  RequestStatus = "pending"
	RequestApproved RequestStatus = "approved"
	RequestRejected RequestStatus = "rejected"
)

type MovementType string

const (
	MovementIn         MovementType = "in"
	MovementOut        MovementType = "out"
	MovementAdjustment MovementType = "adjustment"
)

type ItemRequest struct {
	ID                uuid.UUID     `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	RequesterID       uuid.UUID     `gorm:"type:uuid;not null"`
	Requester         User          `gorm:"foreignKey:RequesterID;constraint:OnDelete:RESTRICT;"`
	ApproverID        *uuid.UUID    `gorm:"type:uuid"` // null until approved/rejected
	Approver          *User         `gorm:"foreignKey:ApproverID;constraint:OnDelete:SET NULL;"`
	ItemID            uuid.UUID     `gorm:"type:uuid;not null"`
	Item              Item          `gorm:"foreignKey:ItemID;constraint:OnDelete:RESTRICT;"`
	QuantityRequested int           `gorm:"not null;check:quantity_requested > 0"`
	Status            RequestStatus `gorm:"type:varchar(50);default:'pending';index"`
	Notes             string        `gorm:"type:text"`
	RequestedAt       time.Time     `gorm:"autoCreateTime"`
	ApprovedAt        *time.Time
}

type StockMovement struct {
	ID           uuid.UUID    `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	ItemID       uuid.UUID    `gorm:"type:uuid;not null"`
	Item         Item         `gorm:"foreignKey:ItemID;constraint:OnDelete:CASCADE;"`
	RequestID    *uuid.UUID   `gorm:"type:uuid"` // null for manual adjustment
	Request      *ItemRequest `gorm:"foreignKey:RequestID;constraint:OnDelete:SET NULL;"`
	QtyChange    int          `gorm:"not null"`
	MovementType MovementType `gorm:"type:varchar(50);not null"`
	MovedAt      time.Time    `gorm:"autoCreateTime"`
}

func (m *ItemRequest) BeforeCreate(tx *gorm.DB) (err error) {
	if m.ID == uuid.Nil {
		m.ID = uuid.New()
	}
	return
}

func (m *StockMovement) BeforeCreate(tx *gorm.DB) (err error) {
	if m.ID == uuid.Nil {
		m.ID = uuid.New()
	}
	return
}
