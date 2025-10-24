package leaderboard_service

import (
	"context"
	"log/slog"
	"main/internal/leaderboard/model/leaderboard"
	"main/internal/pkg/errs"
	"main/internal/pkg/logger"
	"main/internal/team/model/team"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type LeaderboardRepositoryInterface interface {
	Create(context.Context, *gorm.DB, *leaderboard.Leaderboard) error
	UpdateByID(context.Context, *gorm.DB, leaderboard.Leaderboard) error

	GetByTeamId(ctx context.Context, db *gorm.DB, teamId uuid.UUID) (leaderboard.Leaderboard, error)
}

type TeamRepositoryInterface interface {
	GetAcceptTeamsByEventId(context.Context, *gorm.DB, uuid.UUID) ([]team.Team, error)
}

type LeaderboardService struct {
	log                    *slog.Logger
	db                     *gorm.DB
	leaderboardRepostitory LeaderboardRepositoryInterface
	teamRepository         TeamRepositoryInterface
}

func New(log *slog.Logger, storage *gorm.DB, leaderboardRepository LeaderboardRepositoryInterface, teamRepository TeamRepositoryInterface) *LeaderboardService {
	return &LeaderboardService{
		log:                    log,
		db:                     storage,
		leaderboardRepostitory: leaderboardRepository,
		teamRepository:         teamRepository,
	}
}

func (s *LeaderboardService) InitForTeams(ctx context.Context, transactionDB *gorm.DB, eventID uuid.UUID) error {
	const op = "service.LeaderboardService.InitForTeams"

	log := s.log.With(slog.String("op", op))

	teams, err := s.teamRepository.GetAcceptTeamsByEventId(ctx, transactionDB, eventID)
	if err != nil {
		log.Error("failed get accept teams", slog.String("id", eventID.String()), logger.Err(err))
		return errs.ErrInternalError
	}

	for _, team := range teams {
		newLeaderboard := leaderboard.Leaderboard{
			TeamID:  team.ID,
			EventID: eventID,
		}
		err := s.leaderboardRepostitory.Create(ctx, transactionDB, &newLeaderboard)
		if err != nil {
			log.Error("failed create leaderboard", logger.Err(err))
			return errs.ErrInternalError
		}
	}

	return nil
}

func (s *LeaderboardService) AddGameResult(ctx context.Context, teamId uuid.UUID, enemyId uuid.UUID, teamScore int) error {
	const op = "service.LeaderboardService.AddGameResult"

	log := s.log.With(slog.String("op", op))

	err := s.AddGameResultTx(ctx, s.db, teamId, enemyId, teamScore)
	if err != nil {
		log.Error("failed add game result", logger.Err(err))
		return errs.ErrInternalError
	}
	return nil
}

func (s *LeaderboardService) AddGameResultTx(ctx context.Context, transactionDB *gorm.DB, teamId uuid.UUID, enemyId uuid.UUID, teamScore int) error {
	const op = "service.LeaderboardService.AddGameResultTx"

	log := s.log.With(slog.String("op", op))

	leaderboardModel, err := s.leaderboardRepostitory.GetByTeamId(ctx, transactionDB, teamId)
	if err != nil {
		log.Error("failed get leaderboard by teamId", logger.Err(err))
		return err
	}

	if leaderboardModel.EnemiesIds == nil {
		leaderboardModel.EnemiesIds = []uuid.UUID{}
	}
	if leaderboardModel.Scores == nil {
		leaderboardModel.Scores = []int{}
	}

	leaderboardModel.EnemiesIds = append(leaderboardModel.EnemiesIds, enemyId)
	leaderboardModel.Scores = append(leaderboardModel.Scores, teamScore)

	err = s.leaderboardRepostitory.UpdateByID(ctx, transactionDB, leaderboardModel)
	if err != nil {
		log.Error("failed update leaderboard result", logger.Err(err))
		return err
	}

	return nil
}
