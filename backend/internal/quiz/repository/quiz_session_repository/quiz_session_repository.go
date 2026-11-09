package quiz_session_repository

import (
	"context"
	"fmt"
	"main/internal/common/repository"
	"main/internal/quiz/model/quiz_session"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type QuizSessionRepository struct {
	repository.RepositoryBase[quiz_session.QuizSession]
}

func New() *QuizSessionRepository {
	return &QuizSessionRepository{}
}

func (r *QuizSessionRepository) GetOrCreate(ctx context.Context, db *gorm.DB, eventID uuid.UUID) (quiz_session.QuizSession, error) {
	const op = "repository.QuizSessionRepository.GetOrCreate"

	var session quiz_session.QuizSession
	err := db.WithContext(ctx).
		Where(quiz_session.QuizSession{EventID: eventID}).
		FirstOrCreate(&session).
		Error
	if err != nil {
		return quiz_session.QuizSession{}, fmt.Errorf("%s: %w", op, err)
	}
	return session, nil
}

func (r *QuizSessionRepository) Save(ctx context.Context, db *gorm.DB, session quiz_session.QuizSession) error {
	const op = "repository.QuizSessionRepository.Save"

	err := db.WithContext(ctx).Save(&session).Error
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	return nil
}
