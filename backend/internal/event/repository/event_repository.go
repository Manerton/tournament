package event_repository

import (
	"context"
	"fmt"
	"main/internal/event/model/event"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type EventRepository struct {
}

func New() *EventRepository {
	return &EventRepository{}
}

func (r *EventRepository) GetByID(ctx context.Context, db *gorm.DB, id uuid.UUID) (*event.Event, error) {
	const op = "repository.EventRepository.GetByID"

	var model event.Event
	err := db.WithContext(ctx).First(&model, id).Error
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	return &model, nil
}

func (r *EventRepository) GetAll(ctx context.Context, db *gorm.DB) ([]event.Event, error) {
	const op = "repository.EventRepository.GetAll"

	var models []event.Event
	err := db.WithContext(ctx).Find(&models).Error
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	return models, nil
}

func (r *EventRepository) GetByPreviousEventID(ctx context.Context, db *gorm.DB, previousEventID uuid.UUID) ([]event.Event, error) {
	const op = "repository.EventRepository.GetByPreviousEventID"

	var models []event.Event
	err := db.WithContext(ctx).Where(event.Event{PreviousEventID: &previousEventID}).Find(&models).Error
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	return models, nil
}

func (r *EventRepository) Create(ctx context.Context, db *gorm.DB, model event.Event) (uuid.UUID, error) {
	const op = "repository.EventRepository.Create"

	err := db.WithContext(ctx).Create(&model).Error
	if err != nil {
		return uuid.Nil, fmt.Errorf("%s: %w", op, err)
	}
	return model.ID, nil
}

func (r *EventRepository) UpdateByID(ctx context.Context, db *gorm.DB, model event.Event) error {
	const op = "repository.EventRepository.UpdateByID"
	err := db.WithContext(ctx).Updates(&model).Error
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	return nil
}

func (r *EventRepository) DeleteByID(ctx context.Context, db *gorm.DB, id uuid.UUID) error {
	const op = "repository.EventRepository.DeleteByID"
	err := db.WithContext(ctx).Delete(event.Event{}, id).Error
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	return nil
}
