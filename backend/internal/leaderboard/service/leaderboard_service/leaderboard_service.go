package leaderboard_service

import (
	"context"
	"log/slog"
	"main/internal/event/model/event"
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
	GetByPreviousEventID(context.Context, *gorm.DB, uuid.UUID) ([]event.Event, error)
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

func New(
	log *slog.Logger,
	storage *gorm.DB,
	leaderboardRepository LeaderboardRepositoryInterface,
	teamRepository TeamRepositoryInterface,
	teamParticipantRepository TeamParticipantRepositoryInterface,
	eventRepository EventRepositoryInterface,
	gameRepository GameRepositoryInterface,
	gameParticipantsRepository GameParticipantRepositoryInterface,
) *LeaderboardService {
	return &LeaderboardService{
		log:                        log,
		db:                         storage,
		leaderboardRepository:      leaderboardRepository,
		teamRepository:             teamRepository,
		teamParticipantRepository:  teamParticipantRepository,
		eventRepository:            eventRepository,
		GameRepositoryInterface:    gameRepository,
		GameParticipantsRepository: gameParticipantsRepository,
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

func (s *LeaderboardService) GetLeaderboardWins(ctx context.Context, EventID string) ([]leaderboard_dto.LeaderboardWinDTOResponse, error) {
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

	var responses []leaderboard_dto.LeaderboardWinDTOResponse
	for _, lb := range leaderboards {
		responses = append(responses, leaderboard_dto.LeaderboardWinDTOResponse{
			ID:        lb.ID.String(),
			Name:      lb.Team.Name,
			ScoreWins: lb.WinCount,
			Score:     int(calculations.Sum(lb.Scores)),
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

func (s *LeaderboardService) GetLeaderboardScoreLive(ctx context.Context, EventID string) ([]leaderboard_dto.LeaderboardDTOResponse, error) {
	const op = "service.LeaderboardService.GetLeaderboardScoreLive"
	log := s.log.With(slog.String("op", op))

	eventID, err := uuid.Parse(EventID)
	if err != nil {
		log.Error("failed parse id to uuid", slog.String("id", EventID), logger.Err(err))
		return nil, errs.ErrInvalidId
	}

	// 1. Получаем все команды, участвующие в событии
	teams, err := s.teamRepository.GetAcceptTeamsByEventId(ctx, s.db, eventID)
	if err != nil {
		log.Error("failed get teams by event id", logger.Err(err))
		return nil, err
	}

	// Карта участник → команда
	participantToTeam := make(map[uuid.UUID]uuid.UUID)
	teamName := make(map[uuid.UUID]string)

	for _, t := range teams {
		teamName[t.ID] = t.Name

		participants, err := s.teamParticipantRepository.GetAllByTeamIDPreload(ctx, s.db, t.ID)
		if err != nil {
			log.Error("failed get team participants", logger.Err(err))
			return nil, err
		}

		for _, part := range participants {
			participantToTeam[part.ParticipantID] = t.ID
		}
	}

	// 2. Получаем все туры, связанные с событием
	tours, err := s.eventRepository.GetByPreviousEventID(ctx, s.db, eventID)
	if err != nil {
		log.Error("failed get tours", logger.Err(err))
		return nil, err
	}

	// 3. Собираем все игры
	var games []game.Game
	for _, tour := range tours {
		gamesInTour, err := s.GameRepositoryInterface.GetByEventId(ctx, s.db, tour.ID)
		if err != nil {
			log.Error("failed get games", logger.Err(err))
			return nil, err
		}
		games = append(games, gamesInTour...)
	}

	// 4. Считаем очки по участникам, потом суммируем в разрезе команд
	teamScores := make(map[uuid.UUID]int)

	for _, g := range games {
		gameParticipants, err := s.GameParticipantsRepository.GetAllByGameID(ctx, s.db, g.ID)
		if err != nil {
			log.Error("failed get game participants", logger.Err(err))
			return nil, err
		}

		for _, gp := range gameParticipants {
			teamID, ok := participantToTeam[gp.ParticipantID]
			if !ok {
				log.Warn("participant not found in team map", slog.String("participant_id", gp.ParticipantID.String()))
				continue
			}
			teamScores[teamID] += gp.Score
		}
	}

	// 5. Формируем ответ
	var responses []leaderboard_dto.LeaderboardDTOResponse
	for id, score := range teamScores {
		responses = append(responses, leaderboard_dto.LeaderboardDTOResponse{
			ID:    id.String(),
			Name:  teamName[id],
			Score: score,
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

	Teams, err := s.teamRepository.GetAcceptTeamsByEventId(ctx, s.db, eventID)
	if err != nil {
		log.Error("failed get teams by event id", logger.Err(err))
		return nil, err
	}

	participantTeamName := map[uuid.UUID]string{}
	participantName := map[uuid.UUID]string{}

	for _, t := range Teams {

		participants, err := s.teamParticipantRepository.GetAllByTeamIDPreload(ctx, s.db, t.ID)
		if err != nil {
			log.Error("failed get team participants", logger.Err(err))
			return nil, err
		}
		for _, part := range participants {
			participantTeamName[part.ParticipantID] = t.Name
			participantName[part.ParticipantID] = part.Participant.FIO
		}
	}

	Tours, err := s.eventRepository.GetByPreviousEventID(ctx, s.db, eventID)
	if err != nil {
		log.Error("failed get tours", logger.Err(err))
		return nil, err
	}

	var Games []game.Game
	for _, tour := range Tours {
		gameInTour, err := s.GameRepositoryInterface.GetByEventId(ctx, s.db, tour.ID)
		if err != nil {
			log.Error("failed get games", logger.Err(err))
			return nil, err
		}
		Games = append(Games, gameInTour...)
	}

	scores := make(map[uuid.UUID]int)
	for _, g := range Games {
		gameParticipants, err := s.GameParticipantsRepository.GetAllByGameID(ctx, s.db, g.ID)
		if err != nil {
			log.Error("failed get game participants", logger.Err(err))
			return nil, err
		}
		for _, gp := range gameParticipants {
			scores[gp.ParticipantID] += gp.Score
		}
	}

	var responses []leaderboard_dto.LeaderboardPlayesDTOResponse
	for id, score := range scores {
		responses = append(responses, leaderboard_dto.LeaderboardPlayesDTOResponse{
			TeamName: participantTeamName[id],
			Name:     participantName[id],
			Score:    score,
		})
	}

	return responses, nil
}
