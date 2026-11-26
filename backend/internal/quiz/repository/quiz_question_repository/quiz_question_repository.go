package quiz_question_repository

import (
	"context"
	"fmt"
	"main/internal/common/repository"
	"main/internal/quiz/model/quiz_question"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type QuizQuestionRepository struct {
	repository.RepositoryBase[quiz_question.QuizQuestion]
}

func New() *QuizQuestionRepository {
	return &QuizQuestionRepository{}
}

func (r *QuizQuestionRepository) GetByEventAndRound(ctx context.Context, db *gorm.DB, eventID uuid.UUID, round int) ([]quiz_question.QuizQuestion, error) {
	const op = "repository.QuizQuestionRepository.GetByEventAndRound"

	var questions []quiz_question.QuizQuestion
	err := db.WithContext(ctx).
		Where(quiz_question.QuizQuestion{EventID: eventID, Round: round}).
		Order("question_number").
		Find(&questions).Error
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	return questions, nil
}

func (r *QuizQuestionRepository) GetByEvent(ctx context.Context, db *gorm.DB, eventID uuid.UUID) ([]quiz_question.QuizQuestion, error) {
	const op = "repository.QuizQuestionRepository.GetByEvent"

	var questions []quiz_question.QuizQuestion
	err := db.WithContext(ctx).
		Where(quiz_question.QuizQuestion{EventID: eventID}).
		Order("round, question_number").
		Find(&questions).Error
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	return questions, nil
}

func (r *QuizQuestionRepository) GetByNumber(ctx context.Context, db *gorm.DB, eventID uuid.UUID, round int, number int) (quiz_question.QuizQuestion, error) {
	const op = "repository.QuizQuestionRepository.GetByNumber"

	var question quiz_question.QuizQuestion
	err := db.WithContext(ctx).
		Where(quiz_question.QuizQuestion{EventID: eventID, Round: round, QuestionNumber: number}).
		First(&question).Error
	if err != nil {
		return quiz_question.QuizQuestion{}, fmt.Errorf("%s: %w", op, err)
	}
	return question, nil
}

func (r *QuizQuestionRepository) Upsert(ctx context.Context, db *gorm.DB, question quiz_question.QuizQuestion) error {
	const op = "repository.QuizQuestionRepository.Upsert"

	err := db.WithContext(ctx).
		Where(quiz_question.QuizQuestion{
			EventID:        question.EventID,
			Round:          question.Round,
			QuestionNumber: question.QuestionNumber,
		}).
		Assign(question).
		FirstOrCreate(&question).Error
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	return nil
}

// ResetFlagsByEvent clears the transient per-run flags of all questions of the
// event: used (reserve pool, shown once) and cancelled (annulled question).
func (r *QuizQuestionRepository) ResetFlagsByEvent(ctx context.Context, db *gorm.DB, eventID uuid.UUID) error {
	const op = "repository.QuizQuestionRepository.ResetFlagsByEvent"

	err := db.WithContext(ctx).
		Model(&quiz_question.QuizQuestion{}).
		Where(quiz_question.QuizQuestion{EventID: eventID}).
		Updates(map[string]interface{}{"used": false, "cancelled": false}).Error
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	return nil
}
