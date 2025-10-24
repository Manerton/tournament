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
	GetAll(ctx context.Context, db *gorm.DB) ([]event.Event, error)
	GetByID(ctx context.Context, db *gorm.DB, id uuid.UUID) (event.Event, error)
	GetByPreviousEventID(ctx context.Context, db *gorm.DB, previousEventID uuid.UUID) ([]event.Event, error)
	GetByFilter(ctx context.Context, db *gorm.DB, filter event.Event) ([]event.Event, error)

	GetLastTour(ctx context.Context, db *gorm.DB, parentId uuid.UUID) (event.Event, error)
	GetNumLastTour(ctx context.Context, db *gorm.DB, parentId uuid.UUID) (int, error)

	Create(ctx context.Context, db *gorm.DB, _ *event.Event) error
	UpdateByID(ctx context.Context, db *gorm.DB, _ event.Event) error
	DeleteByID(ctx context.Context, db *gorm.DB, id uuid.UUID) error
}

type LeaderboardServiceInterface interface {
	InitForTeams(ctx context.Context, transactionDB *gorm.DB, eventID uuid.UUID) error
}

type GameServiceInterface interface {
	AllGamesIsFinished(ctx context.Context, eventId uuid.UUID) (bool, error)
	CreateFirstGamesByInitialScores(ctx context.Context, mainEventId uuid.UUID, tourEventId uuid.UUID) error
}

type TeamRepositoryInterface interface {
	GetCountTeams(ctx context.Context, db *gorm.DB, eventId uuid.UUID) (int, error)
}

type EventService struct {
	db  *gorm.DB
	log *slog.Logger

	repository     EventRepositoryInterface
	teamRepository TeamRepositoryInterface

	gameService        GameServiceInterface
	leaderboardService LeaderboardServiceInterface
}

func New(log *slog.Logger, db *gorm.DB,
	repository EventRepositoryInterface, teamRepository TeamRepositoryInterface,
	gameService GameServiceInterface, leaderboardService LeaderboardServiceInterface) *EventService {
	return &EventService{
		db:  db,
		log: log,

		repository:     repository,
		teamRepository: teamRepository,

		gameService:        gameService,
		leaderboardService: leaderboardService,
	}
}

func (s *EventService) GetAll(ctx context.Context) ([]event_dto.EventDTOResponse, error) {
	const op = "service.EventService.GetAll"

	log := s.log.With(slog.String("op", op))

	eventsModel, err := s.repository.GetAll(ctx, s.db)
	if err != nil {
		log.Error("failed get all evens", logger.Err(err))
		return nil, errs.ErrInternalError
	}

	return event_mapper.ToDTOEvents(eventsModel), nil
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

func (s *EventService) GetTournamentsByStatus(ctx context.Context, statusStr string) ([]event_dto.EventDTOResponse, error) {
	const op = "service.EventService.GetTournamentsByStatus"

	log := s.log.With(slog.String("op", op))

	status, err := strconv.Atoi(statusStr)
	if err != nil {
		log.Error("failed convert status to int", slog.String("status", statusStr), logger.Err(err))
		return nil, errs.ErrBadRequest
	}

	events, err := s.repository.GetByFilter(ctx, s.db, event.Event{Type: event.EventType, Status: status})
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

func (s *EventService) GetAllTournaments(ctx context.Context) ([]event_dto.EventDTOResponse, error) {
	const op = "service.EventService.GetAllTournaments"

	log := s.log.With(slog.String("op", op))

	models, err := s.repository.GetByFilter(ctx, s.db, event.Event{Type: event.EventType})
	if err != nil {
		log.Error("failed get all tournaments", logger.Err(err))
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

	parentEvent, err := s.repository.GetByID(ctx, transactionDB, parentId)
	if err != nil {
		return uuid.Nil, fmt.Errorf("failed get parent event")
	}

	newTourNum := lastTour.Stage + 1
	if newTourNum >= parentEvent.MaxStage {
		return uuid.Nil, fmt.Errorf("failed create new tour, max tour overhead")
	}

	newTour := event.Event{
		Name:            "Тур" + strconv.Itoa(newTourNum),
		PreviousEventID: &parentId,
		Stage:           newTourNum,
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

func (s *EventService) StartTournament(ctx context.Context, eventId string) error {
	const op = "service.EventService.StartTournament"
	log := s.log.With(slog.String("op", op))

	uid, err := uuid.Parse(eventId)
	if err != nil {
		log.Error("failed parse id to uuid", slog.String("id", eventId), logger.Err(err))
		return errs.ErrInvalidId
	}

	countTeams, err := s.teamRepository.GetCountTeams(ctx, s.db, uid)
	if err != nil {
		log.Error("failed get count teams", logger.Err(err))
		return errs.ErrInternalError
	}

	if countTeams == 0 {
		log.Warn("count teams is 0")
		return errs.ErrStartTournament
	}

	transactionDB := s.db.Begin()
	err = s.leaderboardService.InitForTeams(ctx, transactionDB, uid)
	if err != nil {
		transactionDB.Rollback()
		log.Error("failed init leaderboard", logger.Err(err))
		return errs.ErrInternalError
	}

	tourId, err := s.createTour(ctx, transactionDB, uid)
	if err != nil {
		transactionDB.Rollback()
		log.Error("failed create first tour", logger.Err(err))
		return errs.ErrInternalError
	}

	transactionDB.Commit()
	// Создаём игры
	return s.gameService.CreateFirstGamesByInitialScores(ctx, uid, tourId)
}

func (s *EventService) FinishEvent(ctx context.Context, id string) error {
	const op = "service.EventService.FinishEvent"
	log := s.log.With(slog.String("op", op))

	uid, err := uuid.Parse(id)
	if err != nil {
		log.Error("failed parse id to uuid", slog.String("id", id), logger.Err(err))
		return errs.ErrInvalidId
	}

	eventModel, err := s.repository.GetByID(ctx, s.db, uid)
	if err != nil {
		log.Error("failed get event by id", logger.Err(err))
		return errs.ErrInternalError
	}

	if eventModel.Type == event.StageType {
		return s.finishTour(ctx, eventModel)
	}
	return s.finishTournament(ctx, eventModel)
}

func (s *EventService) finishTournament(ctx context.Context, eventModel event.Event) error {
	const op = "service.EventService.finishTournament"
	log := s.log.With(slog.String("op", op))

	eventChilds, err := s.repository.GetByPreviousEventID(ctx, s.db, eventModel.ID)
	if err != nil {
		log.Error("failed get child events (tours)", logger.Err(err))
		return errs.ErrInternalError
	}

	for _, eventChild := range eventChilds {
		if eventChild.Status == event.ActiveStatus {
			log.Warn("failed finish tour", slog.String("event", eventModel.Name))
			return errs.ErrFinishedEvent.Wrap(fmt.Sprintf("%s: не завершён", eventChild.Name))
		}
	}

	eventModel.Status = event.FinishedStatus
	err = s.repository.UpdateByID(ctx, s.db, eventModel)
	if err != nil {
		log.Error("failed update event status")
		return errs.ErrInternalError
	}

	return nil
}

func (s *EventService) finishTour(ctx context.Context, eventModel event.Event) error {
	const op = "service.EventService.finishTour"
	log := s.log.With(slog.String("op", op))

	ok, err := s.gameService.AllGamesIsFinished(ctx, eventModel.ID)
	if err != nil {
		log.Error("failed check all games is finished", logger.Err(err))
		return errs.ErrInternalError
	}

	if !ok {
		log.Warn("not all game is finished")
		return errs.ErrFinishedEvent.Wrap("не все игры завершены")
	}

	eventModel.Status = event.FinishedStatus
	err = s.repository.UpdateByID(ctx, s.db, eventModel)
	if err != nil {
		log.Error("failed update event", logger.Err(err))
		return errs.ErrInternalError
	}

	return nil
}

func (s *EventService) Update(ctx context.Context, id string, dto event_dto.EventDTOUpdate) error {
	const op = "service.EventService.Update"

	log := s.log.With(slog.String("op", op))
	uid, err := uuid.Parse(id)
	if err != nil {
		log.Error("failed parse id to uuid", slog.String("id", id), logger.Err(err))
		return errs.ErrInvalidId
	}

	updatedModel := event_mapper.ToModelUpdate(uid, dto)
	err = s.repository.UpdateByID(ctx, s.db, updatedModel)
	if err != nil {
		log.Error("failed update event", logger.Err(err))
		return errs.ErrInternalError
	}
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

	err = s.repository.DeleteByID(ctx, s.db, uid)
	if err != nil {
		log.Error("failed delete event", logger.Err(err))
		return errs.ErrInternalError
	}

	return nil
}
