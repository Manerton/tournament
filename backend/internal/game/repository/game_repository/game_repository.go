package game_repository

import (
	"context"
	"fmt"
	"main/internal/common/repository"
	"main/internal/game/model/game"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type GameRespository struct {
	repository.RepositoryBase[game.Game]
}

func New() *GameRespository {
	return &GameRespository{}
}

func (r *GameRespository) GetByEventId(ctx context.Context, db *gorm.DB, eventId uuid.UUID) ([]game.Game, error) {
	const op = "repository.GameRepository.GetByEventId"

	var games []game.Game
	err := db.WithContext(ctx).Preload("TeamA").Preload("TeamB").Where(game.Game{EventID: eventId}).Find(&games).Error
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	return games, nil
}

func (r *GameRespository) GetByIDPreload(ctx context.Context, db *gorm.DB, id uuid.UUID) (game.Game, error) {
	const op = "repository.GameRepository.GetByIDPreload"

	result := game.Game{}
	err := db.WithContext(ctx).Preload("TeamA").Preload("TeamB").Where(game.Game{ID: id}).First(&result).Error
	if err != nil {
		return game.Game{}, fmt.Errorf("%s: %w", op, err)
	}
	return result, nil
}
