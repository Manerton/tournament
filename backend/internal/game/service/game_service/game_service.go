package game_service

import (
	"context"
	"fmt"
	"log/slog"
	"main/internal/game/dto/game_dto"
	"main/internal/game/mapper/game_mapper"
	"main/internal/game/model/game"
	"main/internal/game/model/game_participant"
	"main/internal/leaderboard/model/leaderboard"
	"main/internal/pkg/calculations"
	"main/internal/pkg/errs"
	"main/internal/pkg/logger"
	"main/internal/team/model/team_participant"
	"slices"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type GameRepositoryInterface interface {
	Create(ctx context.Context, db *gorm.DB, _ *game.Game) error
	GetByEventId(ctx context.Context, db *gorm.DB, eventId uuid.UUID) ([]game.Game, error)
	DeleteByID(ctx context.Context, db *gorm.DB, id uuid.UUID) error
}

type GameParticipantRepositoryInterface interface {
	Create(ctx context.Context, db *gorm.DB, _ *game_participant.GameParticipant) error
}

type LeaderboardRepostioryInterface interface {
	GetByEventId(ctx context.Context, db *gorm.DB, eventId uuid.UUID) ([]leaderboard.Leaderboard, error)
}

type TeamParticipantRepositoryInterface interface {
	GetAllByTeamID(ctx context.Context, db *gorm.DB, teamID uuid.UUID) ([]team_participant.TeamParticipant, error)
}

type GameService struct {
	log                       *slog.Logger
	gameRepository            GameRepositoryInterface
	gameParticipantRepository GameParticipantRepositoryInterface
	leaderboardRepostiory     LeaderboardRepostioryInterface
	teamParticipantRepository TeamParticipantRepositoryInterface
	db                        *gorm.DB
}

func New(log *slog.Logger, db *gorm.DB,
	gameRepository GameRepositoryInterface,
	leaderboardRepostiory LeaderboardRepostioryInterface,
	gameParticipantRepository GameParticipantRepositoryInterface,
	teamParticipantRepository TeamParticipantRepositoryInterface) *GameService {
	return &GameService{
		log:                       log,
		db:                        db,
		gameRepository:            gameRepository,
		leaderboardRepostiory:     leaderboardRepostiory,
		gameParticipantRepository: gameParticipantRepository,
		teamParticipantRepository: teamParticipantRepository,
	}
}

func (s *GameService) GetByEventId(ctx context.Context, eventId string) ([]game_dto.GameDTOResponse, error) {
	const op = "service.GameService.GetByEventId"
	log := slog.With(slog.String("op", op))

	uid, err := uuid.Parse(eventId)
	if err != nil {
		log.Error("failed parse id to uuid", slog.String("id", eventId), logger.Err(err))
		return nil, errs.ErrInvalidId
	}

	games, err := s.gameRepository.GetByEventId(ctx, s.db, uid)
	if err != nil {
		log.Error("failed get games by eventId", slog.String("eventId", eventId), logger.Err(err))
		return nil, errs.ErrInternalError
	}

	return game_mapper.ToManyDTOGame(games), nil
}

func (s *GameService) DeleteByID(ctx context.Context, id string) error {
	const op = "service.GameService.DeleteByID"

	log := slog.With(slog.String("op", op))
	uid, err := uuid.Parse(id)
	if err != nil {
		log.Error("failed parse id to uuid", slog.String("id", id), logger.Err(err))
		return errs.ErrBadRequest
	}

	err = s.gameRepository.DeleteByID(ctx, s.db, uid)
	if err != nil {
		log.Error("failed delete game by id", logger.Err(err))
		return errs.ErrInternalError
	}

	return nil
}

// Ссоздание игр по швейцарской системе
func (s *GameService) CreateGames(ctx context.Context, mainEventId uuid.UUID, tourEventId uuid.UUID) error {
	const op = "service.GameService.CreateGames"

	log := slog.With(slog.String("op", op))

	// Получаем историю игр команд
	leaderboards, err := s.leaderboardRepostiory.GetByEventId(ctx, s.db, mainEventId)
	if err != nil {
		log.Error("failed get leaderboards by eventId", slog.String("eventId", mainEventId.String()), logger.Err(err))
		return errs.ErrInternalError
	}

	// Сортируем команды по убыванию побед и очков
	leaderboards = calculations.SortLeaderboards(leaderboards)
	// Создаём пары команд, которые ещё не играли между собой
	err = s.createGamePairs(ctx, leaderboards, tourEventId)
	if err != nil {
		log.Error("failed create game pairs", logger.Err(err))
		return errs.ErrInternalError
	}

	return nil
}

func (s *GameService) createGamePairs(ctx context.Context, leaderboards []leaderboard.Leaderboard, tourEventId uuid.UUID) error {
	const op = "service.GameService.createGamePairs"

	log := slog.With(slog.String("op", op))

	// Проходим по всем командам и создаём игры с командами, с которыми ещё не играли
	for i := 0; i < len(leaderboards); i++ {
		for j := i + 1; j < len(leaderboards); j++ {
			if !slices.Contains(leaderboards[i].EnemieIds, leaderboards[j].TeamID) {
				newGame := &game.Game{
					EventID:  tourEventId,
					TeamA:    leaderboards[i].TeamID,
					TeamB:    leaderboards[j].TeamID,
					Finished: game.NotFinished,
				}
				// Создаём игру
				err := s.createGame(ctx, newGame)
				if err != nil {
					log.Error("failed create game", logger.Err(err))
					return errs.ErrInternalError
				}
			}
		}
	}

	return nil
}

func (s *GameService) createGame(ctx context.Context, game *game.Game) error {
	const op = "service.GameService.createGame"

	// Открываем транзакцию
	transactionBegin := s.db.Begin()
	if transactionBegin.Error != nil {
		return fmt.Errorf("%s: failed begin transaction: %w", op, transactionBegin.Error)
	}

	// создаём игру
	err := s.gameRepository.Create(ctx, transactionBegin, game)
	if err != nil {
		transactionBegin.Rollback()
		return fmt.Errorf("%s: failed create game: %w", op, err)
	}

	// создаём участников игры из состава команд А
	err = s.createGameParticipants(ctx, transactionBegin, game.ID, game.TeamA)
	if err != nil {
		transactionBegin.Rollback()
		return fmt.Errorf("%s: failed create game participants for team A: %w", op, err)
	}
	// создаём участников игры из состава команд B
	err = s.createGameParticipants(ctx, transactionBegin, game.ID, game.TeamB)
	if err != nil {
		transactionBegin.Rollback()
		return fmt.Errorf("%s: failed create game participants for team B: %w", op, err)
	}
	// коммитим транзакцию
	transactionBegin.Commit()
	return nil
}

func (s *GameService) createGameParticipants(ctx context.Context, transactionDB *gorm.DB, gameID uuid.UUID, teamID uuid.UUID) error {
	const op = "service.GameService.createGameParticipants"

	// получение всех участников команды
	participants, err := s.teamParticipantRepository.GetAllByTeamID(ctx, transactionDB, teamID)
	if err != nil {
		return fmt.Errorf("%s: failed get participant %w", op, err)
	}

	for _, p := range participants {
		newGameparticipant := &game_participant.GameParticipant{
			GameID:        gameID,
			ParticipantID: p.ParticipantID,
		}
		// запись участника в игру
		err = s.gameParticipantRepository.Create(ctx, transactionDB, newGameparticipant)
		if err != nil {
			return fmt.Errorf("%s: %w", op, err)
		}
	}

	return nil
}

func (s *GameService) FinishGame(ctx context.Context, gameID uuid.UUID) error {
	const op = "service.GameService.FinishGame"

	return nil
}

func (s *GameService) AllGamesIsFinished(ctx context.Context, eventId uuid.UUID) (bool, error) {
	const op = "service.GameService.AllGamesIsFinished"
	log := slog.With(slog.String("op", op))

	games, err := s.gameRepository.GetByEventId(ctx, s.db, eventId)
	if err != nil {
		log.Error("failed get games by eventId", slog.String("eventId", eventId.String()), logger.Err(err))
		return false, errs.ErrInternalError
	}

	for _, g := range games {
		if g.Finished == game.NotFinished {
			return false, nil
		}
	}

	return true, nil
}
