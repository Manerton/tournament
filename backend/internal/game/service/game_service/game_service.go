package game_service

import (
	"context"
	"log/slog"
	"main/internal/game/model/game"
	"main/internal/pkg/errs"
	"main/internal/pkg/logger"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type GameRepositoryInterface interface {
	Create(ctx context.Context, db *gorm.DB, _ *game.Game) error
	GetByEventId(ctx context.Context, db *gorm.DB, eventId uuid.UUID) ([]game.Game, error)
	DeleteByID(ctx context.Context, db *gorm.DB, id uuid.UUID) error
}

type GameService struct {
	log        *slog.Logger
	repository GameRepositoryInterface
	db         *gorm.DB
}

func New(log *slog.Logger, db *gorm.DB, repository GameRepositoryInterface) *GameService {
	return &GameService{
		log:        log,
		db:         db,
		repository: repository,
	}
}

func (s *GameService) DeleteByID(ctx context.Context, id string) error {
	const op = "service.GameService.DeleteByID"

	log := slog.With(slog.String("op", op))
	uid, err := uuid.Parse(id)
	if err != nil {
		log.Error("failed parse id to uuid", slog.String("id", id), logger.Err(err))
		return errs.ErrBadRequest
	}

	err = s.repository.DeleteByID(ctx, s.db, uid)
	if err != nil {
		log.Error("failed delete game by id", logger.Err(err))
		return errs.ErrInternalError
	}

	return nil
}
