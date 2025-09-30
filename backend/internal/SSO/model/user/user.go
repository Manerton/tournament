package user

import "github.com/google/uuid"

type User struct {
	ID           uuid.UUID `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	Login        string    `gorm:"type:varchar(255);not null;uniqueIndex"`
	Password     string    `gorm:"type:varchar(255);not null"`
}
