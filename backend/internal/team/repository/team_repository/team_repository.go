package team_repository

import (
	"context"
	"fmt"
	"main/internal/common/repository"
	"main/internal/team/model/team"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type TeamRepository struct {
	repository.RepositoryBase[team.Team]
}

func New() *TeamRepository {
	return &TeamRepository{}
}

func (r *TeamRepository) GetByEventID(ctx context.Context, db *gorm.DB, id uuid.UUID) ([]team.Team, error) {
	const op = "repostiory.TeamReposiory.GetByEventID"

	var result []team.Team
	err := db.WithContext(ctx).Where(team.Team{EventID: id}).Find(&result).Error
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	return result, nil
}

func (r *TeamRepository) GetAcceptTeamsByEventId(ctx context.Context, db *gorm.DB, id uuid.UUID) ([]team.Team, error) {
	const op = "repostiory.TeamReposiory.GetAcceptTeamsByEventId"

	var result []team.Team
	err := db.WithContext(ctx).Where(team.Team{EventID: id, Status: team.AcceptStatus}).Find(&result).Error
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	return result, nil
}

// Кол-во принятых команд
func (r *TeamRepository) GetCountTeams(ctx context.Context, db *gorm.DB, eventId uuid.UUID) (int, error) {
	const op = "repository.TeamRepository.GetCountTeams"

	var result int64
	err := db.WithContext(ctx).Model(team.Team{}).Where(team.Team{EventID: eventId, Status: team.AcceptStatus}).Count(&result).Error
	if err != nil {
		return 0, fmt.Errorf("%s: %w", op, err)
	}
	return int(result), nil
}
