package event_service

import (
	"context"
	"main/internal/event/dto/event_dto"
	"main/internal/event/event_mapper"
	"main/internal/event/model/event"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type EventRepositoryInterface interface {
	Create(ctx context.Context, db *gorm.DB, _ event.Event) (uuid.UUID, error)
	UpdateByID(ctx context.Context, db *gorm.DB, _ event.Event) error
	GetByID(ctx context.Context, db *gorm.DB, id uuid.UUID) (*event.Event, error)
	GetAll(ctx context.Context, db *gorm.DB) ([]event.Event, error)
	GetByPreviousEventID(ctx context.Context, db *gorm.DB, previousEventID uuid.UUID) ([]event.Event, error)
}

type EventService struct {
	db         *gorm.DB
	repository EventRepositoryInterface
}

func New(db *gorm.DB, repository EventRepositoryInterface) *EventService {
	return &EventService{
		db:         db,
		repository: repository,
	}
}

func (s *EventService) Create(ctx context.Context, dto event_dto.EventDTOCreate) (uuid.UUID, error) {
	const op = "service.EventService.Create"

	model := event_mapper.ToModelEvent(dto)
	id, err := s.repository.Create(ctx, s.db, model)
	if err != nil {
		return uuid.Nil, err
	}

	return id, nil
}
