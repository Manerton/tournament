package user_repository

import (
	"context"
	"fmt"

	"main/internal/SSO/model/user"
	"main/internal/common/repository"

	"gorm.io/gorm"
)

type userRepository struct {
	repository.RepositoryBase[user.User]
}

func New() *userRepository {
	return &userRepository{}
}

// Получить пользователя по логину
func (r *userRepository) GetByLogin(ctx context.Context, db *gorm.DB, login string) (user.User, error) {
	const op = "repositories.UserRepository.GetByLogin"
	var entity user.User
	if err := db.WithContext(ctx).First(&entity, "login = ?", login).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return user.User{}, nil
		}
		return user.User{}, fmt.Errorf("%s: %w", op, err)
	}
	return entity, nil
}

// Обновить пользователя
func (r *userRepository) Update(ctx context.Context, db *gorm.DB, u user.User) error {
	const op = "repositories.UserRepository.Update"
	if err := db.WithContext(ctx).Save(u).Error; err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	return nil
}
