package team

import (
	"main/internal/event/model/event"

	"github.com/google/uuid"
)

type Team struct {
	ID      uuid.UUID   `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	Name    string      `gorm:"type:varchar(255);not null;uniqueIndex"`
	EventID uuid.UUID   `gorm:"type:uuid;not null;index"`
	Event   event.Event `gorm:"foreignKey:EventID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
}
