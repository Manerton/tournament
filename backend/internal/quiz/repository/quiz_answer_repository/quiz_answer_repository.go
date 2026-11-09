package quiz_answer_repository

import (
	"context"
	"fmt"
	"main/internal/common/repository"
	"main/internal/quiz/model/quiz_answer"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type QuizAnswerRepository struct {
	repository.RepositoryBase[quiz_answer.QuizAnswer]
}

func New() *QuizAnswerRepository {
	return &QuizAnswerRepository{}
}

func (r *QuizAnswerRepository) GetByTeamAndQuestion(ctx context.Context, db *gorm.DB, eventID uuid.UUID, round int, questionNumber int, teamID uuid.UUID) (quiz_answer.QuizAnswer, error) {
	const op = "repository.QuizAnswerRepository.GetByTeamAndQuestion"

	var answer quiz_answer.QuizAnswer
	err := db.WithContext(ctx).
		Where(quiz_answer.QuizAnswer{
			EventID:        eventID,
			Round:          round,
			QuestionNumber: questionNumber,
			TeamID:         teamID,
		}).
		First(&answer).Error
	if err != nil {
		return quiz_answer.QuizAnswer{}, fmt.Errorf("%s: %w", op, err)
	}
	return answer, nil
}

func (r *QuizAnswerRepository) GetByTeamAndRound(ctx context.Context, db *gorm.DB, eventID uuid.UUID, round int, teamID uuid.UUID) ([]quiz_answer.QuizAnswer, error) {
	const op = "repository.QuizAnswerRepository.GetByTeamAndRound"

	var answers []quiz_answer.QuizAnswer
	err := db.WithContext(ctx).
		Where(quiz_answer.QuizAnswer{EventID: eventID, Round: round, TeamID: teamID}).
		Find(&answers).Error
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	return answers, nil
}

func (r *QuizAnswerRepository) Upsert(ctx context.Context, db *gorm.DB, answer quiz_answer.QuizAnswer) error {
	const op = "repository.QuizAnswerRepository.Upsert"

	err := db.WithContext(ctx).
		Where(quiz_answer.QuizAnswer{
			EventID:        answer.EventID,
			Round:          answer.Round,
			QuestionNumber: answer.QuestionNumber,
			TeamID:         answer.TeamID,
		}).
		Assign(answer).
		FirstOrCreate(&answer).Error
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	return nil
}
