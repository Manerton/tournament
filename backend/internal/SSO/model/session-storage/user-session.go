package session_storage

import (
	"time"

	"github.com/google/uuid"
)

type UserSession struct {
	ID           uuid.UUID `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	UserId       uuid.UUID `gorm:"type:uuid;not null;index"`
	SessionToken string    `gorm:"type:varchar(255);not null;uniqueIndex"`
	CSRFToken    string    `gorm:"type:varchar(255);not null"`

	CreatedAt time.Time `gorm:"not null;default:now()"`
	UpdatedAt time.Time `gorm:"not null;default:now()"`

	ExpiresAt time.Time `gorm:"not null;index"` // принудительное истечение в БД
}
