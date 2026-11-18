package quiz_button_repository

import (
	"context"
	"fmt"
	"main/internal/common/repository"
	"main/internal/quiz/model/quiz_button"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type QuizButtonRepository struct {
	repository.RepositoryBase[quiz_button.QuizButton]
}

func New() *QuizButtonRepository {
	return &QuizButtonRepository{}
}

func (r *QuizButtonRepository) GetByID(ctx context.Context, db *gorm.DB, id string) (quiz_button.QuizButton, error) {
	const op = "repository.QuizButtonRepository.GetByID"

	var button quiz_button.QuizButton
	err := db.WithContext(ctx).
		Preload("QuizTeam").
		Where(quiz_button.QuizButton{ID: id}).
		First(&button).Error
	if err != nil {
		return quiz_button.QuizButton{}, fmt.Errorf("%s: %w", op, err)
	}
	return button, nil
}

func (r *QuizButtonRepository) Upsert(ctx context.Context, db *gorm.DB, button quiz_button.QuizButton) error {
	const op = "repository.QuizButtonRepository.Upsert"

	err := db.WithContext(ctx).
		Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "id"}},
			DoUpdates: clause.AssignmentColumns([]string{"quiz_team_id"}),
		}).
		Create(&button).Error
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	return nil
}

func (r *QuizButtonRepository) Delete(ctx context.Context, db *gorm.DB, id string) error {
	const op = "repository.QuizButtonRepository.Delete"

	err := db.WithContext(ctx).
		Where(quiz_button.QuizButton{ID: id}).
		Delete(&quiz_button.QuizButton{}).Error
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	return nil
}

// ListByEvent returns all buttons of the event with their quiz teams preloaded.
func (r *QuizButtonRepository) ListByEvent(ctx context.Context, db *gorm.DB, eventID uuid.UUID) ([]quiz_button.QuizButton, error) {
	const op = "repository.QuizButtonRepository.ListByEvent"

	var buttons []quiz_button.QuizButton
	err := db.WithContext(ctx).
		Joins("JOIN quiz_teams ON quiz_teams.id = quiz_buttons.quiz_team_id").
		Preload("QuizTeam").
		Preload("QuizTeam.Team").
		Where("quiz_teams.event_id = ?", eventID).
		Order("quiz_buttons.id").
		Find(&buttons).Error
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	return buttons, nil
}
