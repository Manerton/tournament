package event_repository

import (
	"context"
	"fmt"
	"main/internal/common/repository"
	"main/internal/event/model/event"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type EventRepository struct {
	repository.RepositoryBase[event.Event]
}

func New() *EventRepository {
	return &EventRepository{}
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
	err := db.WithContext(ctx).Where(event.Event{PreviousEventID: &previousEventID}).Order("stage ASC").Find(&models).Error
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	return models, nil
}

func (r *EventRepository) GetLastTour(ctx context.Context, db *gorm.DB, parentId uuid.UUID) (event.Event, error) {
	const op = "repository.EventRepository.GetLastTour"
	var lastTour event.Event
	err := db.Model(&event.Event{}).Where(event.Event{PreviousEventID: &parentId}).Order("stage DESC").Limit(1).First(&lastTour).Error
	if err != nil {
		return event.Event{}, fmt.Errorf("%s: %w", op, err)
	}
	return lastTour, nil
}

func (r *EventRepository) GetNumLastTour(ctx context.Context, db *gorm.DB, parentId uuid.UUID) (int, error) {
	const op = "repository.EventRepository.GetNumLastTour"
	var stage int
	err := db.WithContext(ctx).Model(&event.Event{}).
		Where("previous_event_id = ?", parentId).
		Select("COALESCE(MAX(stage), 0)").
		Scan(&stage).Error
	if err != nil {
		return 0, fmt.Errorf("%s: %w", op, err)
	}
	return stage, nil
}

func (r *EventRepository) GetByPreviousEventIDAndStage(ctx context.Context, db *gorm.DB, parentId uuid.UUID, stage int) (event.Event, error) {
	const op = "repository.EventRepository.GetByPreviousEventIDAndStage"
	var result = event.Event{}
	err := db.WithContext(ctx).Model(&event.Event{}).Where(event.Event{PreviousEventID: &parentId, Stage: stage}).Find(&result).Error
	if err != nil {
		return event.Event{}, fmt.Errorf("%s: %w", op, err)
	}
	return result, nil
}

func (r *EventRepository) UpdateMaxStageById(ctx context.Context, db *gorm.DB, id uuid.UUID, maxStage int) error {
	const op = "repository.EventRepository.UpdateMaxStageById"
	err := db.WithContext(ctx).Model(event.Event{}).Where(event.Event{ID: id}).UpdateColumn("max_stage", maxStage).Error
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	return nil

}
