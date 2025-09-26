package team_participant

import (
	"main/internal/participant/model/participant"
	"main/internal/team/model/team"

	"github.com/google/uuid"
)

type TeamParticipant struct {
	ID            uuid.UUID               `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	TeamID        uuid.UUID               `gorm:"type:uuid"`
	Team          team.Team               `gorm:"foreignKey:TeamID;constraint:OnUpdate:CASCADE;OnDelete:CASCADE;"`
	ParticipantID uuid.UUID               `gorm:"type:uuid"`
	Participant   participant.Participant `gorm:"foreignKey:ParticipantID;constraint:OnUpdate:CASCADE;OnDelete:CASCADE;"`
}
