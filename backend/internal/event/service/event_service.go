package event_service

import (
	"context"
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

	GetNumLastTour(ctx context.Context, db *gorm.DB, parentId uuid.UUID) (int, error)

	DeleteByID(ctx context.Context, id uuid.UUID) error
}

type EventService struct {
	db         *gorm.DB
	repository EventRepositoryInterface
	log        *slog.Logger
}

func New(db *gorm.DB, repository EventRepositoryInterface) *EventService {
	return &EventService{
		db:         db,
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

	err := s.repository.Create(ctx, transactionDB, &model)
	if err != nil {
		transactionDB.Rollback()
		log.Error("failed create event", logger.Err(err))
		return uuid.Nil, errs.ErrInternalError
	}

	// if not main event
	if model.PreviousEventID != nil {
		err := s.createTourWithGames(ctx, transactionDB, model.ID)
		if err != nil {
			transactionDB.Rollback()
			log.Error("failed create tour with games", logger.Err(err))
			return uuid.Nil, errs.ErrInternalError
		}
	}

	return model.ID, nil
}

func (s *EventService) createTourWithGames(ctx context.Context, transactionDB *gorm.DB, parentId uuid.UUID) error {
	// create tour
	lastTourNum, err := s.repository.GetNumLastTour(ctx, transactionDB, parentId)
	if err != nil {
		return err
	}

	lastTourNum += 1
	newTour := event.Event{
		Name:            "Тур" + strconv.Itoa(lastTourNum),
		PreviousEventID: &parentId,
		Stage:           lastTourNum,
	}

	err = s.repository.Create(ctx, transactionDB, &newTour)
	if err != nil {
		return err
	}
	_ = newTour.ID

	// TODO! Create games

	return nil
}

func (s *EventService) Delete(ctx context.Context, id string) error {
	const op = "service.EventService.Delete"

	log := s.log.With(slog.String("op", op))

	uid, err := uuid.Parse(id)
	if err != nil {
		log.Error("failed parse id to uuid", slog.String("id", id), logger.Err(err))
		return errs.ErrInvalidId
	}

	err = s.repository.DeleteByID(ctx, uid)
	if err != nil {
		log.Error("failed delete event", logger.Err(err))
		return errs.ErrInternalError
	}

	return nil
}
