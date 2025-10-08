package game

import (
	"main/internal/event/model/event"

	"github.com/google/uuid"
)

const (
	NotFinished = 3
	TeamAWon    = 1
	TeamBWon    = 2
)

type Game struct {
	ID       uuid.UUID   `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	TeamA    uuid.UUID   `gorm:"type:uuid"`
	TeamB    uuid.UUID   `gorm:"type:uuid"`
	EventID  uuid.UUID   `gorm:"type:uuid"`
	Event    event.Event `gorm:"foreignKey:EventID;constraint:OnUpdate:CASCADE;OnDelete:CASCADE;"`
	Finished int         `gorm:"type:int;not null;default:3"` // 3 - not finished, 1 - teamA won, 2 - teamB won
}
