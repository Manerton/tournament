package quiz_buzz_repository

import (
	"context"
	"fmt"
	"main/internal/common/repository"
	"main/internal/quiz/model/quiz_buzz"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type QuizBuzzRepository struct {
	repository.RepositoryBase[quiz_buzz.QuizBuzz]
}

func New() *QuizBuzzRepository {
	return &QuizBuzzRepository{}
}

func (r *QuizBuzzRepository) GetByQuestion(ctx context.Context, db *gorm.DB, eventID uuid.UUID, round int, questionNumber int) ([]quiz_buzz.QuizBuzz, error) {
	const op = "repository.QuizBuzzRepository.GetByQuestion"

	var buzzes []quiz_buzz.QuizBuzz
	err := db.WithContext(ctx).
		Where(quiz_buzz.QuizBuzz{EventID: eventID, Round: round, QuestionNumber: questionNumber}).
		Order("seq").
		Find(&buzzes).Error
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	return buzzes, nil
}

func (r *QuizBuzzRepository) GetByTeamAndQuestion(ctx context.Context, db *gorm.DB, eventID uuid.UUID, round int, questionNumber int, teamID uuid.UUID) (quiz_buzz.QuizBuzz, error) {
	const op = "repository.QuizBuzzRepository.GetByTeamAndQuestion"

	var buzz quiz_buzz.QuizBuzz
	err := db.WithContext(ctx).
		Where(quiz_buzz.QuizBuzz{
			EventID:        eventID,
			Round:          round,
			QuestionNumber: questionNumber,
			TeamID:         teamID,
		}).
		First(&buzz).Error
	if err != nil {
		return quiz_buzz.QuizBuzz{}, fmt.Errorf("%s: %w", op, err)
	}
	return buzz, nil
}

func (r *QuizBuzzRepository) Save(ctx context.Context, db *gorm.DB, buzz quiz_buzz.QuizBuzz) error {
	const op = "repository.QuizBuzzRepository.Save"

	err := db.WithContext(ctx).Save(&buzz).Error
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	return nil
}

func (r *QuizBuzzRepository) DeleteByQuestion(ctx context.Context, db *gorm.DB, eventID uuid.UUID, round int, questionNumber int) error {
	const op = "repository.QuizBuzzRepository.DeleteByQuestion"

	err := db.WithContext(ctx).
		Where(quiz_buzz.QuizBuzz{EventID: eventID, Round: round, QuestionNumber: questionNumber}).
		Delete(&quiz_buzz.QuizBuzz{}).Error
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	return nil
}
