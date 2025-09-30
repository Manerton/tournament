package session_repository

import (
	"context"
	"fmt"

	session_storage "main/internal/SSO/model/session-storage"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type SessionRepository struct{}

func NewSessionRepository() *SessionRepository {
	return &SessionRepository{}
}

// Создать новую сессию
func (r *SessionRepository) Create(ctx context.Context, db *gorm.DB, entity *session_storage.UserSession) error {
	const op = "repositories.SessionRepository.Create"
	if err := db.WithContext(ctx).Create(entity).Error; err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	return nil
}

// Получить сессию по ID
func (r *SessionRepository) GetByID(ctx context.Context, db *gorm.DB, id uuid.UUID) (*session_storage.UserSession, error) {
	const op = "repositories.SessionRepository.GetByID"
	var entity session_storage.UserSession
	if err := db.WithContext(ctx).First(&entity, "id = ?", id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	return &entity, nil
}

// Получить сессию по UserID
func (r *SessionRepository) GetByUserID(ctx context.Context, db *gorm.DB, userID uuid.UUID) (*session_storage.UserSession, error) {
	const op = "repositories.SessionRepository.GetByUserID"
	var entity session_storage.UserSession
	if err := db.WithContext(ctx).First(&entity, "user_id = ?", userID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	return &entity, nil
}

// Получить сессию по токену
func (r *SessionRepository) GetByToken(ctx context.Context, db *gorm.DB, token string) (*session_storage.UserSession, error) {
	const op = "repositories.SessionRepository.GetByToken"
	var entity session_storage.UserSession
	if err := db.WithContext(ctx).First(&entity, "session_token = ?", token).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	return &entity, nil
}

// Обновить сессию
func (r *SessionRepository) Update(ctx context.Context, db *gorm.DB, entity *session_storage.UserSession) error {
	const op = "repositories.SessionRepository.Update"
	if err := db.WithContext(ctx).Save(entity).Error; err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	return nil
}

// Удалить сессию по ID
func (r *SessionRepository) Delete(ctx context.Context, db *gorm.DB, id uuid.UUID) error {
	const op = "repositories.SessionRepository.Delete"
	if err := db.WithContext(ctx).Delete(&session_storage.UserSession{}, "id = ?", id).Error; err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	return nil
}
