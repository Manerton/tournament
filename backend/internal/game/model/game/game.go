package game

import "github.com/google/uuid"

type Game struct {
	ID   uuid.UUID `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	Name string    `gorm:"type:varchar(255);not null;uniqueIndex"`
}
