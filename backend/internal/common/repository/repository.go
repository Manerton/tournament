package repository

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type RepositoryBase[T any] struct {
}

func (r *RepositoryBase[T]) GetAll(ctx context.Context, db *gorm.DB) ([]T, error) {
	const op = "repository.RepositoryBase.GetAll"

	var model []T
	err := db.WithContext(ctx).Find(&model).Error
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	return model, nil
}

func (r *RepositoryBase[T]) GetByID(ctx context.Context, db *gorm.DB, id uuid.UUID) (T, error) {
	const op = "repository.RepositoryBase.GetByID"

	var model T
	err := db.WithContext(ctx).First(&model, id).Error
	if err != nil {
		return model, fmt.Errorf("%s: %w", op, err)
	}
	return model, nil
}

func (r *RepositoryBase[T]) GetByFilter(ctx context.Context, db *gorm.DB, filter T) ([]T, error) {
	const op = "repository.RepositoryBase.GetByID"

	var model []T
	err := db.WithContext(ctx).Where(filter).Find(&model).Error
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	return model, nil
}

func (r *RepositoryBase[T]) Create(ctx context.Context, db *gorm.DB, model *T) error {
	const op = "repository.RepositoryBase.Create"

	err := db.WithContext(ctx).Create(&model).Error
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	return nil
}

func (r *RepositoryBase[T]) UpdateByID(ctx context.Context, db *gorm.DB, model T) error {
	const op = "repository.RepositoryBase.UpdateByID"

	err := db.WithContext(ctx).Updates(model).Error
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	return nil
}

func (r *RepositoryBase[T]) DeleteByID(ctx context.Context, db *gorm.DB, id uuid.UUID) error {
	const op = "repository.RepositoryBase.DeleteByID"

	var model = new(T)
	err := db.WithContext(ctx).Delete(model, id).Error
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	return nil
}
