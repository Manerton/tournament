package event

import (
	"time"

	"github.com/google/uuid"
)

const (
	KnockoutDraw   = 1
	RoundRobinDraw = 2
	SwissDraw      = 3
	QuizDraw       = 4
)

const (
	EventType = 1
	StageType = 2
)

const (
	ActiveStatus   = 1
	FinishedStatus = 2
	RegisterStatus = 3
)

type Event struct {
	ID              uuid.UUID  `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	Name            string     `gorm:"type:varchar(255);not null;uniqueIndex"`
	StartDate       time.Time  `gorm:"type:date;not null;index"`
	EndDate         time.Time  `gorm:"type:date;not null;index"`
	PreviousEventID *uuid.UUID `gorm:"type:uuid;index"`
	Type            int
	Status          int
	PreviousEvent   *Event `gorm:"foreignKey:PreviousEventID;constraint:OnUpdate:CASCADE,OnDelete:SET NULL;"`
	TypeDraws       int    `gorm:"type:varchar(100);not null;index"`
	Stage           int
	MaxStage        int
}
