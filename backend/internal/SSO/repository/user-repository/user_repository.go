package user_repository

import (
	"context"
	"fmt"

	"main/internal/SSO/model/user"

	"github.com/google/uuid"
	"gorm.io/gorm"
)


type userRepository struct{}

// Создание пользователя
func (r *userRepository) Create(ctx context.Context, db *gorm.DB, entity *user.User) error {
	const op = "repositories.UserRepository.Create"
	if err := db.WithContext(ctx).Create(entity).Error; err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	return nil
}

// Получить пользователя по ID
func (r *userRepository) GetByID(ctx context.Context, db *gorm.DB, id uuid.UUID) (*user.User, error) {
	const op = "repositories.UserRepository.GetByID"
	var entity user.User
	if err := db.WithContext(ctx).First(&entity, "id = ?", id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	return &entity, nil
}

// Получить пользователя по логину
func (r *userRepository) GetByLogin(ctx context.Context, db *gorm.DB, login string) (*user.User, error) {
	const op = "repositories.UserRepository.GetByLogin"
	var entity user.User
	if err := db.WithContext(ctx).First(&entity, "login = ?", login).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	return &entity, nil
}

// Получить всех пользователей
func (r *userRepository) GetAll(ctx context.Context, db *gorm.DB) ([]user.User, error) {
	const op = "repositories.UserRepository.GetAll"
	var entities []user.User
	if err := db.WithContext(ctx).Find(&entities).Error; err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	return entities, nil
}

// Обновить пользователя
func (r *userRepository) Update(ctx context.Context, db *gorm.DB, entity *user.User) error {
	const op = "repositories.UserRepository.Update"
	if err := db.WithContext(ctx).Save(entity).Error; err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	return nil
}

// Удалить пользователя
func (r *userRepository) Delete(ctx context.Context, db *gorm.DB, id uuid.UUID) error {
	const op = "repositories.UserRepository.Delete"
	if err := db.WithContext(ctx).Delete(&user.User{}, "id = ?", id).Error; err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	return nil
}
