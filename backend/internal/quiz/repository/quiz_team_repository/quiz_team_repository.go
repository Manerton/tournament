package quiz_team_repository

import (
	"context"
	"fmt"
	"main/internal/common/repository"
	"main/internal/quiz/model/quiz_team"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type QuizTeamRepository struct {
	repository.RepositoryBase[quiz_team.QuizTeam]
}

func New() *QuizTeamRepository {
	return &QuizTeamRepository{}
}

func (r *QuizTeamRepository) GetByEventID(ctx context.Context, db *gorm.DB, eventID uuid.UUID) ([]quiz_team.QuizTeam, error) {
	const op = "repository.QuizTeamRepository.GetByEventID"

	var teams []quiz_team.QuizTeam
	err := db.WithContext(ctx).
		Preload("Team").
		Where(quiz_team.QuizTeam{EventID: eventID}).
		Order("access_code").
		Find(&teams).Error
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	return teams, nil
}

func (r *QuizTeamRepository) GetByTeamID(ctx context.Context, db *gorm.DB, eventID uuid.UUID, teamID uuid.UUID) (quiz_team.QuizTeam, error) {
	const op = "repository.QuizTeamRepository.GetByTeamID"

	var team quiz_team.QuizTeam
	err := db.WithContext(ctx).
		Preload("Team").
		Where(quiz_team.QuizTeam{EventID: eventID, TeamID: teamID}).
		First(&team).Error
	if err != nil {
		return quiz_team.QuizTeam{}, fmt.Errorf("%s: %w", op, err)
	}
	return team, nil
}

func (r *QuizTeamRepository) GetByAccessCode(ctx context.Context, db *gorm.DB, eventID uuid.UUID, code string) (quiz_team.QuizTeam, error) {
	const op = "repository.QuizTeamRepository.GetByAccessCode"

	var team quiz_team.QuizTeam
	err := db.WithContext(ctx).
		Preload("Team").
		Where(quiz_team.QuizTeam{EventID: eventID, AccessCode: code}).
		First(&team).Error
	if err != nil {
		return quiz_team.QuizTeam{}, fmt.Errorf("%s: %w", op, err)
	}
	return team, nil
}

func (r *QuizTeamRepository) DeleteByEventAndTeamID(ctx context.Context, db *gorm.DB, eventID uuid.UUID, teamID uuid.UUID) error {
	const op = "repository.QuizTeamRepository.DeleteByEventAndTeamID"

	err := db.WithContext(ctx).
		Where(quiz_team.QuizTeam{EventID: eventID, TeamID: teamID}).
		Delete(&quiz_team.QuizTeam{}).Error
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	return nil
}

func (r *QuizTeamRepository) Save(ctx context.Context, db *gorm.DB, team quiz_team.QuizTeam) error {
	const op = "repository.QuizTeamRepository.Save"

	err := db.WithContext(ctx).Save(&team).Error
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	return nil
}
