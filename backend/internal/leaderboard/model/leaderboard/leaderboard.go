package leaderboard

import (
	"encoding/json"
	"main/internal/event/model/event"
	"main/internal/team/model/team"

	"github.com/google/uuid"
)

type Leaderboard struct {
	ID      uuid.UUID       `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	TeamID  uuid.UUID       `gorm:"type:uuid"`
	Team    team.Team       `gorm:"foreignKey:TeamID;constraint:OnUpdate:CASCADE;OnDelete:CASCADE;"`
	EventID uuid.UUID       `gorm:"type:uuid"`
	Event   event.Event     `gorm:"foreignKey:EventID;constraint:OnUpdate:CASCADE;OnDelete:CASCADE;"`
	Result  json.RawMessage `gorm:"type:jsonb;not null"` // TODO: SOME SHEET
}
