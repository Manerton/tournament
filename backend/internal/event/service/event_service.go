package event_service

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"main/internal/event/dto/event_dto"
	"main/internal/event/event_mapper"
	"main/internal/event/model/event"
	"main/internal/pkg/errs"
	"main/internal/pkg/logger"
	"strconv"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type EventRepositoryInterface interface {
	Create(ctx context.Context, db *gorm.DB, _ *event.Event) error
	UpdateByID(ctx context.Context, db *gorm.DB, _ event.Event) error
	GetByID(ctx context.Context, db *gorm.DB, id uuid.UUID) (event.Event, error)
	GetByPreviousEventID(ctx context.Context, db *gorm.DB, previousEventID uuid.UUID) ([]event.Event, error)
	GetByFilter(ctx context.Context, db *gorm.DB, filter event.Event) ([]event.Event, error)

	GetLastTour(ctx context.Context, db *gorm.DB, parentId uuid.UUID) (event.Event, error)
	GetNumLastTour(ctx context.Context, db *gorm.DB, parentId uuid.UUID) (int, error)

	DeleteByID(ctx context.Context, db *gorm.DB, id uuid.UUID) error
}

type EventService struct {
	db         *gorm.DB
	repository EventRepositoryInterface
	log        *slog.Logger
}

func New(repository EventRepositoryInterface, log *slog.Logger, db *gorm.DB) *EventService {
	return &EventService{
		db:         db,
		log:        log,
		repository: repository,
	}
}

func (s *EventService) GetByID(ctx context.Context, id string) (event_dto.EventDTOResponse, error) {
	const op = "service.EventService.GetByID"

	log := s.log.With(slog.String("op", op))

	uid, err := uuid.Parse(id)
	if err != nil {
		log.Error("failed parse id to uuid", slog.String("id", id), logger.Err(err))
		return event_dto.EventDTOResponse{}, errs.ErrInvalidId
	}

	model, err := s.repository.GetByID(ctx, s.db, uid)
	if err != nil {
		log.Error("failed get event by id", logger.Err(err))
		return event_dto.EventDTOResponse{}, errs.ErrInternalError
	}

	resultDTO := event_mapper.ToDTOEvent(model)
	return resultDTO, nil
}

func (s *EventService) GetActualEvents(ctx context.Context) ([]event_dto.EventDTOResponse, error) {
	const op = "service.EventService.GetActualEvents"

	log := s.log.With(slog.String("op", op))
	events, err := s.repository.GetByFilter(ctx, s.db, event.Event{Type: event.EventType, Status: event.ActiveStatus})
	if err != nil {
		log.Error("failed get actual events", logger.Err(err))
		return nil, errs.ErrInternalError
	}
	return event_mapper.ToDTOEvents(events), nil
}

func (s *EventService) GetByPreviousEventID(ctx context.Context, id string) ([]event_dto.EventDTOResponse, error) {
	const op = "service.EventService.GetByPreviousEventID"

	log := s.log.With(slog.String("op", op))
	uid, err := uuid.Parse(id)
	if err != nil {
		log.Error("failed parse id to uuid", slog.String("id", id), logger.Err(err))
		return nil, errs.ErrInvalidId
	}

	models, err := s.repository.GetByPreviousEventID(ctx, s.db, uid)
	if err != nil {
		log.Error("failed get child events (tour)", logger.Err(err))
		return nil, errs.ErrInternalError
	}

	return event_mapper.ToDTOEvents(models), nil
}

func (s *EventService) Create(ctx context.Context, dto event_dto.EventDTOCreate) (uuid.UUID, error) {
	const op = "service.EventService.Create"

	log := s.log.With(slog.String("op", op))

	model := event_mapper.ToModelEvent(dto)

	transactionDB := s.db.Begin()
	defer transactionDB.Commit()

	var createdId uuid.UUID
	// if not main event
	if model.PreviousEventID != nil {
		uid, err := s.createTour(ctx, transactionDB, *model.PreviousEventID)
		if err != nil {
			transactionDB.Rollback()
			log.Error("failed create tour with games", logger.Err(err))
			return uuid.Nil, errs.ErrInternalError
		}
		createdId = uid
	} else {
		model.Type = event.EventType
		// simple create main event
		err := s.repository.Create(ctx, transactionDB, &model)
		if err != nil {
			transactionDB.Rollback()
			log.Error("failed create event", logger.Err(err))
			return uuid.Nil, errs.ErrInternalError
		}
		createdId = model.ID
	}

	return createdId, nil
}

func (s *EventService) createTour(ctx context.Context, transactionDB *gorm.DB, parentId uuid.UUID) (uuid.UUID, error) {
	// get last tour
	lastTour, err := s.repository.GetLastTour(ctx, transactionDB, parentId)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		// если просто ошибка
		return uuid.Nil, err
	}

	if errors.Is(err, gorm.ErrRecordNotFound) {
		// если запись не найдена то первый тур
		parentEvent, err := s.repository.GetByID(ctx, transactionDB, parentId)
		if err != nil {
			return uuid.Nil, err
		}
		newTour := event.Event{
			Name:            "Тур 1",
			PreviousEventID: &parentId,
			Stage:           1,
			Status:          event.ActiveStatus,
			Type:            event.StageType,
			StartDate:       parentEvent.StartDate,
			EndDate:         parentEvent.EndDate,
			TypeDraws:       parentEvent.TypeDraws,
		}
		err = s.repository.Create(ctx, transactionDB, &newTour)
		if err != nil {
			return uuid.Nil, err
		}
		return newTour.ID, nil
	}
	// если не первый тур
	if lastTour.Status != event.FinishedStatus {
		// если последний тур не завершён
		return uuid.Nil, fmt.Errorf("failed create new tour: last tour not finished")
	}

	lastTourNum := lastTour.Stage + 1
	newTour := event.Event{
		Name:            "Тур" + strconv.Itoa(lastTourNum),
		PreviousEventID: &parentId,
		Stage:           lastTourNum,
		Status:          event.ActiveStatus,
		StartDate:       lastTour.StartDate,
		EndDate:         lastTour.EndDate,
		Type:            lastTour.Type,
	}
	err = s.repository.Create(ctx, transactionDB, &newTour)
	if err != nil {
		return uuid.Nil, nil
	}

	return newTour.ID, nil
}

func (s *EventService) Delete(ctx context.Context, id string) error {
	const op = "service.EventService.Delete"

	log := s.log.With(slog.String("op", op))

	uid, err := uuid.Parse(id)
	if err != nil {
		log.Error("failed parse id to uuid", slog.String("id", id), logger.Err(err))
		return errs.ErrInvalidId
	}

	err = s.repository.DeleteByID(ctx, s.db, uid)
	if err != nil {
		log.Error("failed delete event", logger.Err(err))
		return errs.ErrInternalError
	}

	return nil
}
