package LiveScoreByTour

import (
	"main/internal/event/model/event"
	"main/internal/participant/model/participant"
	"main/internal/team/model/team"

	"github.com/google/uuid"
)

type LiveScoreByTour struct {
	ID            uuid.UUID               `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	TeamID        uuid.UUID               `gorm:"type:uuid;not null;index:idx_team_event_participant,unique"`
	Team          team.Team               `gorm:"foreignKey:TeamID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
	EventID       uuid.UUID               `gorm:"type:uuid;not null;index:idx_team_event_participant,unique"`
	Event         event.Event             `gorm:"foreignKey:EventID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
	ParticipantID uuid.UUID               `gorm:"type:uuid;not null;index:idx_team_event_participant,unique"`
	Participant   participant.Participant `gorm:"foreignKey:ParticipantID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
	Score         int                     `gorm:"type:int;not null;default:0;check:score >= 0"`
}
