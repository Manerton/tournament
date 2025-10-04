package leaderboard_repository

import (
	"context"
	"fmt"
	"main/internal/common/repository"
	"main/internal/leaderboard/model/leaderboard"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type LeaderboardRepostiory struct {
	repository.RepositoryBase[leaderboard.Leaderboard]
}

func New() *LeaderboardRepostiory {
	return &LeaderboardRepostiory{}
}

func (r *LeaderboardRepostiory) GetByEventId(ctx context.Context, db *gorm.DB, eventId uuid.UUID) ([]leaderboard.Leaderboard, error) {
	const op = "repository.LeaderboardRepository.GetByEventId"
	var leaderboards []leaderboard.Leaderboard
	err := db.WithContext(ctx).Where(leaderboard.Leaderboard{EventID: eventId}).Find(&leaderboards).Error
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	return leaderboards, nil
}
