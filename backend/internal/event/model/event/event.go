package event

import (
	"time"

	"github.com/google/uuid"
)

type Event struct {
	ID              uuid.UUID  `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	Name            string     `gorm:"type:varchar(255);not null;uniqueIndex"`
	StartDate       time.Time  `gorm:"type:date;not null;index"`
	EndDate         time.Time  `gorm:"type:date;not null;index"`
	PreviousEventID *uuid.UUID `gorm:"type:uuid;index"`
	PreviousEvent   *Event     `gorm:"foreignKey:PreviousEventID;constraint:OnUpdate:CASCADE,OnDelete:SET NULL;"`
	Type            string     `gorm:"type:varchar(100);not null;index"` // TODO: Define enum type
}
