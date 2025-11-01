package leaderboard_service

import (
	"context"
	"log/slog"
	"main/internal/game/model/game"
	"main/internal/game/model/game_participant"
	leaderboard_dto "main/internal/leaderboard/dto"
	"main/internal/leaderboard/model/leaderboard"
	"main/internal/pkg/calculations"
	"main/internal/pkg/errs"
	"main/internal/pkg/logger"
	"main/internal/team/model/team"
	"main/internal/team/model/team_participant"

	"github.com/google/uuid"
	"github.com/lib/pq"
	"gorm.io/gorm"
)

type LeaderboardRepositoryInterface interface {
	Create(context.Context, *gorm.DB, *leaderboard.Leaderboard) error
	UpdateByID(context.Context, *gorm.DB, leaderboard.Leaderboard) error
	GetByTeamId(ctx context.Context, db *gorm.DB, teamId uuid.UUID) (leaderboard.Leaderboard, error)
	GetByEventId(ctx context.Context, db *gorm.DB, eventId uuid.UUID) ([]leaderboard.Leaderboard, error)
}

type EventRepositoryInterface interface {
	GetByPreviousEventID(context.Context, *gorm.DB, uuid.UUID) ([]team.Team, error)
}

type TeamRepositoryInterface interface {
	GetAcceptTeamsByEventId(context.Context, *gorm.DB, uuid.UUID) ([]team.Team, error)
}

type TeamParticipantRepositoryInterface interface {
	GetAllByTeamIDPreload(ctx context.Context, db *gorm.DB, teamID uuid.UUID) ([]team_participant.TeamParticipant, error)
}

type GameRepositoryInterface interface {
	GetByEventId(ctx context.Context, db *gorm.DB, eventId uuid.UUID) ([]game.Game, error)
}

type GameParticipantRepositoryInterface interface {
	GetAllByGameID(ctx context.Context, db *gorm.DB, gameID uuid.UUID) ([]game_participant.GameParticipant, error)
}

// type ParticipantRepositoryInterface interface {
// 	GetById(context.Context, *gorm.DB, uuid.UUID) ([]team.Team, error)
// }

type LeaderboardService struct {
	log                        *slog.Logger
	db                         *gorm.DB
	leaderboardRepository      LeaderboardRepositoryInterface
	teamParticipantsRepository TeamParticipantRepositoryInterface
	GameRepositoryInterface    GameRepositoryInterface
	GameParticipantsRepository GameParticipantRepositoryInterface
	teamRepository             TeamRepositoryInterface
	teamParticipantRepository  TeamParticipantRepositoryInterface
	eventRepository            EventRepositoryInterface
	// participantRepository  ParticipantRepositoryInterface
}

func New(log *slog.Logger, storage *gorm.DB, leaderboardRepository LeaderboardRepositoryInterface, teamRepository TeamRepositoryInterface) *LeaderboardService {
	return &LeaderboardService{
		log:                   log,
		db:                    storage,
		leaderboardRepository: leaderboardRepository,
		teamRepository:        teamRepository,
		//participantRepository:  participantRepository,
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
		err := s.leaderboardRepository.Create(ctx, transactionDB, &newLeaderboard)
		if err != nil {
			log.Error("failed create leaderboard", logger.Err(err))
			return errs.ErrInternalError
		}
	}

	return nil
}

func (s *LeaderboardService) AddGameResult(ctx context.Context, teamId uuid.UUID, enemyId uuid.UUID, teamScore int, winA bool) error {
	const op = "service.LeaderboardService.AddGameResult"

	log := s.log.With(slog.String("op", op))

	err := s.AddGameResultTx(ctx, s.db, teamId, enemyId, teamScore, winA)
	if err != nil {
		log.Error("failed add game result", logger.Err(err))
		return errs.ErrInternalError
	}
	return nil
}

func (s *LeaderboardService) AddGameResultTx(ctx context.Context, transactionDB *gorm.DB, teamId uuid.UUID, enemyId uuid.UUID, teamScore int, winA bool) error {
	const op = "service.LeaderboardService.AddGameResultTx"

	log := s.log.With(slog.String("op", op))

	leaderboardModel, err := s.leaderboardRepository.GetByTeamId(ctx, transactionDB, teamId)
	if err != nil {
		log.Error("failed get leaderboard by teamId", logger.Err(err))
		return err
	}

	if leaderboardModel.EnemiesIds == nil {
		leaderboardModel.EnemiesIds = pq.StringArray{}
	}
	if leaderboardModel.Scores == nil {
		leaderboardModel.Scores = pq.Int32Array{}
	}

	if winA {
		leaderboardModel.WinCount += 1
	}

	leaderboardModel.EnemiesIds = append(leaderboardModel.EnemiesIds, enemyId.String())
	leaderboardModel.Scores = append(leaderboardModel.Scores, int32(teamScore))

	err = s.leaderboardRepository.UpdateByID(ctx, transactionDB, leaderboardModel)
	if err != nil {
		log.Error("failed update leaderboard result", logger.Err(err))
		return err
	}

	return nil
}

func (s *LeaderboardService) GetLeaderboardWins(ctx context.Context, EventID string) ([]leaderboard_dto.LeaderboardDTOResponse, error) {
	const op = "service.LeaderboardService.GetLeaderboardWins"

	log := s.log.With(slog.String("op", op))

	eventID, err := uuid.Parse(EventID)
	if err != nil {
		log.Error("failed parse id to uuid", slog.String("id", EventID), logger.Err(err))
		return nil, errs.ErrInvalidId
	}

	leaderboards, err := s.leaderboardRepository.GetByEventId(ctx, s.db, eventID)
	if err != nil {
		log.Error("failed get leaderboards", logger.Err(err))
		return nil, err
	}

	var responses []leaderboard_dto.LeaderboardDTOResponse
	for _, lb := range leaderboards {
		responses = append(responses, leaderboard_dto.LeaderboardDTOResponse{
			ID:    lb.ID.String(),
			Name:  lb.Team.Name,
			Score: lb.WinCount,
		})
	}

	return responses, nil
}

func (s *LeaderboardService) GetLeaderboardScore(ctx context.Context, EventID string) ([]leaderboard_dto.LeaderboardDTOResponse, error) {
	const op = "service.LeaderboardService.GetLeaderboardScore"

	log := s.log.With(slog.String("op", op))

	eventID, err := uuid.Parse(EventID)
	if err != nil {
		log.Error("failed parse id to uuid", slog.String("id", EventID), logger.Err(err))
		return nil, errs.ErrInvalidId
	}

	leaderboards, err := s.leaderboardRepository.GetByEventId(ctx, s.db, eventID)
	if err != nil {
		log.Error("failed get leaderboards", logger.Err(err))
		return nil, err
	}

	var responses []leaderboard_dto.LeaderboardDTOResponse
	for _, lb := range leaderboards {
		responses = append(responses, leaderboard_dto.LeaderboardDTOResponse{
			ID:    lb.ID.String(),
			Name:  lb.Team.Name,
			Score: int(calculations.Sum(lb.Scores)),
		})
	}

	return responses, nil

}

func (s *LeaderboardService) GetParticipantsScore(ctx context.Context, EventID string) ([]leaderboard_dto.LeaderboardPlayesDTOResponse, error) {
	const op = "service.LeaderboardService.GetParticipantsScore"

	log := s.log.With(slog.String("op", op))

	eventID, err := uuid.Parse(EventID)
	if err != nil {
		log.Error("failed parse id to uuid", slog.String("id", EventID), logger.Err(err))
		return nil, errs.ErrInvalidId
	}
	//1 Собираем все команды учавствующие в событии
	Teams, err := s.teamRepository.GetAcceptTeamsByEventId(ctx, s.db, eventID)
	//2 Для каждой команды собираем участников
	participantTeamName := map[uuid.UUID]string{}
	participantName := map[uuid.UUID]string{}

	//TeamParticipants := make(map[uuid.UUID][]team_participant.TeamParticipant)

	for _, Team := range Teams {
		Participants, err := s.teamParticipantRepository.GetAllByTeamIDPreload(ctx, s.db, Team.ID)
		if err != nil {
			log.Error("failed get team participants", logger.Err(err))
			return nil, err
		}
		//TeamParticipants[Team.ID] = Participants
		for _, part := range Participants {
			participantTeamName[part.ParticipantID] = Team.Name
			participantName[part.ParticipantID] = part.Participant.FIO
		}
	}

	//3 Собираем все туры связанные с событием
	Tours, err := s.eventRepository.GetByPreviousEventID(ctx, s.db, eventID)

	//4 Для каждого тура собираем игры
	Games := []game.Game{}

	for _, Tour := range Tours {
		GameinTour, err := s.GameRepositoryInterface.GetByEventId(ctx, s.db, Tour.ID)
		if err != nil {
			log.Error("failed get Games", logger.Err(err))
			return nil, err
		}
		Games = append(GameinTour)
	}

	//5 Считаем очки для каждого участника за все игры
	GameIds := []uuid.UUID{}
	for _, Game := range Games {
		GameIds = append(GameIds, Game.ID)
	}

	scores := map[uuid.UUID]int{}

	for _, GameId := range GameIds {
		GameParticipants, err := s.GameParticipantsRepository.GetAllByGameID(ctx, s.db, GameId)
		if err != nil {
			log.Error("failed get GameParticipants", logger.Err(err))
			return nil, err
		}
		for _, GP := range GameParticipants {
			scores[GP.ParticipantID] += GP.Score
		}

	}

	var responses []leaderboard_dto.LeaderboardPlayesDTOResponse
	for key, score := range scores {
		responses = append(responses, leaderboard_dto.LeaderboardPlayesDTOResponse{
			TeamName: participantTeamName[key],
			Name:     participantName[key],
			Score:    score,
		})
	}

	return responses, nil

}
