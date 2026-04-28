package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type ItemStatus string

const (
	StatusTersedia    ItemStatus = "tersedia"
	StatusDipakai     ItemStatus = "dipakai"
	StatusRusak       ItemStatus = "rusak"
	StatusMaintenance ItemStatus = "maintenance"
)

type Warehouse struct {
	ID          uuid.UUID `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	Name        string    `gorm:"type:varchar(255);not null"`
	Location    string    `gorm:"type:varchar(255)"`
	Description string    `gorm:"type:text"`
	IsActive    bool      `gorm:"default:true"`
	CreatedAt   time.Time
}

type Category struct {
	ID          uuid.UUID `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	Name        string    `gorm:"type:varchar(255);not null"`
	Description string    `gorm:"type:text"`
}

type Item struct {
	ID          uuid.UUID  `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	WarehouseID uuid.UUID  `gorm:"type:uuid;index"`
	Warehouse   Warehouse  `gorm:"foreignKey:WarehouseID;constraint:OnDelete:RESTRICT;"`
	CategoryID  *uuid.UUID `gorm:"type:uuid;index"` // optional
	Category    *Category  `gorm:"foreignKey:CategoryID;constraint:OnDelete:SET NULL;"`
	Name        string     `gorm:"type:varchar(255);not null"`
	SKU         string     `gorm:"type:varchar(100);unique;not null"`
	Quantity    int        `gorm:"not null;default:0"`
	MinStock    int        `gorm:"not null;default:5"`
	Status      ItemStatus `gorm:"type:varchar(50);default:'tersedia'"`
	ImageURL    string     `gorm:"type:text"`
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

func (m *Warehouse) BeforeCreate(tx *gorm.DB) (err error) {
	if m.ID == uuid.Nil {
		m.ID = uuid.New()
	}
	return
}

func (m *Category) BeforeCreate(tx *gorm.DB) (err error) {
	if m.ID == uuid.Nil {
		m.ID = uuid.New()
	}
	return
}

func (m *Item) BeforeCreate(tx *gorm.DB) (err error) {
	if m.ID == uuid.Nil {
		m.ID = uuid.New()
	}
	return
}
