package quiz_tiebreak_team_repository

import (
	"context"
	"fmt"
	"main/internal/common/repository"
	"main/internal/quiz/model/quiz_tiebreak_team"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type QuizTiebreakTeamRepository struct {
	repository.RepositoryBase[quiz_tiebreak_team.QuizTiebreakTeam]
}

func New() *QuizTiebreakTeamRepository {
	return &QuizTiebreakTeamRepository{}
}

func (r *QuizTiebreakTeamRepository) GetByEvent(ctx context.Context, db *gorm.DB, eventID uuid.UUID) ([]quiz_tiebreak_team.QuizTiebreakTeam, error) {
	const op = "repository.QuizTiebreakTeamRepository.GetByEvent"

	var rows []quiz_tiebreak_team.QuizTiebreakTeam
	err := db.WithContext(ctx).
		Where(quiz_tiebreak_team.QuizTiebreakTeam{EventID: eventID}).
		Find(&rows).Error
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	return rows, nil
}

func (r *QuizTiebreakTeamRepository) Save(ctx context.Context, db *gorm.DB, row quiz_tiebreak_team.QuizTiebreakTeam) error {
	const op = "repository.QuizTiebreakTeamRepository.Save"

	err := db.WithContext(ctx).
		Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "event_id"}, {Name: "team_id"}},
			DoUpdates: clause.AssignmentColumns([]string{"won"}),
		}).
		Create(&row).Error
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	return nil
}

func (r *QuizTiebreakTeamRepository) DeleteByEvent(ctx context.Context, db *gorm.DB, eventID uuid.UUID) error {
	const op = "repository.QuizTiebreakTeamRepository.DeleteByEvent"

	err := db.WithContext(ctx).
		Where(quiz_tiebreak_team.QuizTiebreakTeam{EventID: eventID}).
		Delete(&quiz_tiebreak_team.QuizTiebreakTeam{}).Error
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	return nil
}
