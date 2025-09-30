package event

import (
	"time"

	"github.com/google/uuid"
)

const (
	Knockout   = 1
	RoundRobin = 2
	Swiss      = 3
)

type Event struct {
	ID              uuid.UUID  `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	Name            string     `gorm:"type:varchar(255);not null;uniqueIndex"`
	StartDate       time.Time  `gorm:"type:date;not null;index"`
	EndDate         time.Time  `gorm:"type:date;not null;index"`
	PreviousEventID *uuid.UUID `gorm:"type:uuid;index"`
	PreviousEvent   *Event     `gorm:"foreignKey:PreviousEventID;constraint:OnUpdate:CASCADE,OnDelete:SET NULL;"`
	TypeDraws       int        `gorm:"type:varchar(100);not null;index"` // TODO: Define enum type
	Stage           int
}
