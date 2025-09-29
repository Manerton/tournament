package location_repository

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type LocationRepositoryBase[T any] struct {
}

func (r *LocationRepositoryBase[T]) GetByID(ctx context.Context, db *gorm.DB, id uuid.UUID, model T) error {
	const op = "repository.LocationRepository.GetByID"

	err := db.WithContext(ctx).First(model, id).Error
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	return nil
}

func (r *LocationRepositoryBase[T]) Create(ctx context.Context, db *gorm.DB, model T) error {
	const op = "repository.LocationRepository.Create"

	err := db.WithContext(ctx).Create(&model).Error
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	return nil
}

func (r *LocationRepositoryBase[T]) UpdateByID(ctx context.Context, db *gorm.DB, model T) error {
	const op = "repository.LocationRepository.UpdateByID"

	err := db.WithContext(ctx).Updates(model).Error
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	return nil
}

func (r *LocationRepositoryBase[T]) DeleteByID(ctx context.Context, db *gorm.DB, id uuid.UUID) error {
	const op = "repository.LocationRepository.DeleteByID"

	var model = new(T)
	err := db.WithContext(ctx).Delete(model, id).Error
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	return nil
}
