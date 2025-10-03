package game_participant

import (
	"main/internal/game/model/game"
	"main/internal/participant/model/participant"

	"github.com/google/uuid"
)

type GameParticipant struct {
	ID            uuid.UUID               `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	GameID        uuid.UUID               `gorm:"type:uuid"`
	Game          game.Game               `gorm:"foreignKey:GameID;constraint:OnUpdate:CASCADE;OnDelete:CASCADE;"`
	ParticipantID uuid.UUID               `gorm:"type:uuid"`
	Participant   participant.Participant `gorm:"foreignKey:ParticipantID;constraint:OnUpdate:CASCADE;OnDelete:CASCADE;"`
	Score         int                     `gorm:"type:int;not null;default:0;check:score >= 0"`
}
