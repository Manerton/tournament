package session_repository

import (
	"context"
	"fmt"
	"time"

	session_storage "main/internal/SSO/model/session-storage"
	"main/internal/common/repository"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type SessionRepository struct {
	repository.RepositoryBase[session_storage.UserSession]
}

func New() *SessionRepository {
	return &SessionRepository{}
}

// Получить сессию по UserID
func (r *SessionRepository) GetByUserID(ctx context.Context, db *gorm.DB, userID uuid.UUID) (session_storage.UserSession, error) {
	const op = "repositories.SessionRepository.GetByUserID"
	var entity session_storage.UserSession
	if err := db.WithContext(ctx).First(&entity, "user_id = ?", userID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return session_storage.UserSession{}, nil
		}
		return session_storage.UserSession{}, fmt.Errorf("%s: %w", op, err)
	}
	return entity, nil
}

// Получить сессию по токену
func (r *SessionRepository) GetByToken(ctx context.Context, db *gorm.DB, token string) (session_storage.UserSession, error) {
	const op = "repositories.SessionRepository.GetByToken"
	var entity session_storage.UserSession
	if err := db.WithContext(ctx).First(&entity, "session_token = ?", token).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return session_storage.UserSession{}, nil
		}
		return session_storage.UserSession{}, fmt.Errorf("%s: %w", op, err)
	}
	return entity, nil
}

// Обновить сессию
func (r *SessionRepository) Update(ctx context.Context, db *gorm.DB, entity session_storage.UserSession) error {
	const op = "repositories.SessionRepository.Update"
	if err := db.WithContext(ctx).Save(entity).Error; err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	return nil
}

func (r *SessionRepository) DeleteByToken(ctx context.Context, db *gorm.DB, token string) error {
	const op = "repositories.SessionRepository.DeleteByToken"
	if err := db.WithContext(ctx).Delete(&session_storage.UserSession{}, "session_token = ?", token).Error; err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	return nil
}

func (r *SessionRepository) DeleteExpiredSessions(ctx context.Context, db *gorm.DB) error {
	const op = "repositories.SessionRepository.deleteExpiredSessions"
	if err := db.WithContext(ctx).Where("expires_at < ?", time.Now()).Delete(&session_storage.UserSession{}).Error; err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	return nil
}
