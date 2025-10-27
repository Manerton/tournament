package game

import (
	"main/internal/event/model/event"
	"main/internal/team/model/team"

	"github.com/google/uuid"
)

const (
	NotFinished = 3
	TeamAWon    = 1
	TeamBWon    = 2
)

type Game struct {
	ID       uuid.UUID   `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	TeamAId  uuid.UUID   `gorm:"type:uuid;not null" json:"team_a_id"`
	TeamA    team.Team   `gorm:"foreignKey:TeamAId;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"team_a"`
	TeamBId  uuid.UUID   `gorm:"type:uuid;not null" json:"team_b_id"`
	TeamB    team.Team   `gorm:"foreignKey:TeamBId;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"team_b"`
	EventID  uuid.UUID   `gorm:"type:uuid"`
	Event    event.Event `gorm:"foreignKey:EventID;constraint:OnUpdate:CASCADE;OnDelete:CASCADE;"`
	Finished int         `gorm:"type:int;not null;default:3"` // 3 - not finished, 1 - teamA won, 2 - teamB won
}
