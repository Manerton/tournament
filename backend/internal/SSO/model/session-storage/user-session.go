package session_storage

import "github.com/google/uuid"

type UserSession struct {
	ID           uuid.UUID `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	UserId       uuid.UUID `gorm:"type:varchar(255);not null;uniqueIndex"`
	SessionToken string    `gorm:"type:varchar(255);not null"`
	CSRFToken    string    `gorm:"type:varchar(255);not null"`
}
