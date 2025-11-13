package team_participant_repository

import (
	"context"
	"fmt"
	"main/internal/common/repository"
	"main/internal/team/model/team_participant"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type TeamParticipantRepository struct {
	repository.RepositoryBase[team_participant.TeamParticipant]
}

func New() *TeamParticipantRepository {
	return &TeamParticipantRepository{}
}

func (r *TeamParticipantRepository) GetAllByTeamID(ctx context.Context, db *gorm.DB, teamID uuid.UUID) ([]team_participant.TeamParticipant, error) {
	const op = "repository.TeamParticipantRepository.GetAllByTeamID"
	var TeamParticipants []team_participant.TeamParticipant
	err := db.WithContext(ctx).Where(team_participant.TeamParticipant{TeamID: teamID}, teamID).Find(&TeamParticipants).Error
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	return TeamParticipants, nil
}

func (r *TeamParticipantRepository) GetAllByTeamIDPreload(ctx context.Context, db *gorm.DB, teamID uuid.UUID) ([]team_participant.TeamParticipant, error) {
	const op = "repository.TeamParticipantRepository.GetAllByTeamID"
	var TeamParticipants []team_participant.TeamParticipant
	err := db.WithContext(ctx).Preload("Team").Preload("Participant.School").Where(team_participant.TeamParticipant{TeamID: teamID}, teamID).Find(&TeamParticipants).Error
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	return TeamParticipants, nil
}
