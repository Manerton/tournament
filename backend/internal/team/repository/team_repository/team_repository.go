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
