package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type AuditLog struct {
	ID        uuid.UUID    `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	UserID    *uuid.UUID   `gorm:"type:uuid"`
	User      *User        `gorm:"foreignKey:UserID;constraint:OnDelete:SET NULL;"`
	RequestID *uuid.UUID   `gorm:"type:uuid"`
	Request   *ItemRequest `gorm:"foreignKey:RequestID;constraint:OnDelete:SET NULL;"`
	Action    string       `gorm:"type:varchar(255);not null"`
	Payload   string       `gorm:"type:jsonb"` // using jsonb for PostgreSQL
	IPAddress string       `gorm:"type:varchar(45)"`
	CreatedAt time.Time    `gorm:"autoCreateTime;index"`
}

type Notification struct {
	ID      uuid.UUID `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	UserID  uuid.UUID `gorm:"type:uuid;not null"`
	User    User      `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE;"`
	Type    string    `gorm:"type:varchar(50);not null"`
	Message string    `gorm:"type:text;not null"`
	IsRead  bool      `gorm:"default:false"`
	SentAt  time.Time `gorm:"autoCreateTime"`
}

func (m *AuditLog) BeforeCreate(tx *gorm.DB) (err error) {
	if m.ID == uuid.Nil {
		m.ID = uuid.New()
	}
	return
}

func (m *Notification) BeforeCreate(tx *gorm.DB) (err error) {
	if m.ID == uuid.Nil {
		m.ID = uuid.New()
	}
	return
}
