package team_service

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	participant_dto "main/internal/participant/dto"
	"main/internal/participant/model/participant"
	"main/internal/pkg/errs"
	"main/internal/pkg/logger"
	"main/internal/team/dto/team_dto"
	"main/internal/team/model/team"
	"main/internal/team/model/team_participant"
	"main/internal/team/team_mapper"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
	"gorm.io/gorm"
)

type TeamRepositoryInterface interface {
	GetByID(context.Context, *gorm.DB, uuid.UUID) (team.Team, error)
	GetByEventID(context.Context, *gorm.DB, uuid.UUID) ([]team.Team, error)
	GetAcceptTeamsByEventId(context.Context, *gorm.DB, uuid.UUID) ([]team.Team, error)

	Create(context.Context, *gorm.DB, *team.Team) error
	UpdateByID(context.Context, *gorm.DB, team.Team) error
}

type TeamParticipantRepositoryInterface interface {
	Create(context.Context, *gorm.DB, *team_participant.TeamParticipant) error
	GetAllByTeamID(context.Context, *gorm.DB, uuid.UUID) ([]team_participant.TeamParticipant, error)
}

type ParticipantRepositoryInterface interface {
	GetByID(context.Context, *gorm.DB, uuid.UUID) (participant.Participant, error)

	Create(context.Context, *gorm.DB, *participant.Participant) error
}

type TeamService struct {
	teamRepository            TeamRepositoryInterface
	participantRepository     ParticipantRepositoryInterface
	teamParticipantRepository TeamParticipantRepositoryInterface
	log                       *slog.Logger
	db                        *gorm.DB
}

func New(teamRepository TeamRepositoryInterface,
	participantRepository ParticipantRepositoryInterface,
	teamParticipantRepository TeamParticipantRepositoryInterface,
	log *slog.Logger, db *gorm.DB) *TeamService {
	return &TeamService{
		teamRepository:            teamRepository,
		participantRepository:     participantRepository,
		teamParticipantRepository: teamParticipantRepository,
		log:                       log,
		db:                        db,
	}
}

func (s *TeamService) GetByID(ctx context.Context, id string) (team_dto.TeamResponseDTO, error) {
	const op = "service.TeamService.GetByID"
	log := s.log.With(slog.String("op", op))

	uid, err := uuid.Parse(id)
	if err != nil {
		log.Error("failed parse id to uuid", slog.String("id", id), logger.Err(err))
		return team_dto.TeamResponseDTO{}, errs.ErrInvalidId
	}

	teamModel, err := s.teamRepository.GetByID(ctx, s.db, uid)
	if err != nil {
		log.Error("failed get team by id", logger.Err(err))
		return team_dto.TeamResponseDTO{}, errs.ErrInternalError
	}

	return team_mapper.ToTeamDTO(teamModel), nil
}

func (s *TeamService) GetTeamsByEventId(ctx context.Context, eventId string) ([]team_dto.TeamResponseDTO, error) {
	const op = "service.TeamService.GetTeamsByEventId"

	log := s.log.With(slog.String("op", op))

	uid, err := uuid.Parse(eventId)
	if err != nil {
		log.Error("failed parse id to uuid", slog.String("id", eventId), logger.Err(err))
		return nil, errs.ErrInvalidId
	}

	models, err := s.teamRepository.GetByEventID(ctx, s.db, uid)
	if err != nil {
		log.Error("failed get teams by event id", logger.Err(err))
		return nil, errs.ErrInternalError
	}
	return team_mapper.ToTeamsDTO(models), nil
}

func (s *TeamService) GetAcceptTeamsByEventId(ctx context.Context, eventId string) ([]team_dto.TeamResponseDTO, error) {
	const op = "service.TeamService.GetAcceptTeamsByEventId"

	log := s.log.With(slog.String("op", op))

	uid, err := uuid.Parse(eventId)
	if err != nil {
		log.Error("failed parse id to uuid", slog.String("id", eventId), logger.Err(err))
		return nil, errs.ErrInvalidId
	}

	models, err := s.teamRepository.GetAcceptTeamsByEventId(ctx, s.db, uid)
	if err != nil {
		log.Error("failed get accept teams by event id", logger.Err(err))
		return nil, errs.ErrInternalError
	}
	return team_mapper.ToTeamsDTO(models), nil
}

func (s *TeamService) UpdateTeam(ctx context.Context, dto team_dto.TeamUpdateDTO, id string) error {
	const op = "service.TeamService.UpdateTeam"

	log := s.log.With(slog.String("op", op))

	uid, err := uuid.Parse(id)
	if err != nil {
		log.Error("failed parse id to uuid", slog.String("id", id), logger.Err(err))
		return errs.ErrBadRequest
	}

	teamModel := team_mapper.UpdateToModel(dto, uid)
	err = s.teamRepository.UpdateByID(ctx, s.db, teamModel)
	if err != nil {
		log.Error("failed update team", logger.Err(err))
		return errs.ErrInternalError
	}
	return nil
}

func (s *TeamService) RegisterTeam(ctx context.Context, dto team_dto.RegisterTeamRequestDTO) error {
	const op = "service.TeamService.RegisterTeam"

	log := s.log.With(slog.String("op", op))

	teamModel, err := team_mapper.ToTeamModel(dto.TeamCreateDTO)
	if err != nil {
		log.Error("failed to map dto to model", slog.Any("dto", dto), logger.Err(err))
		return errs.ErrBadRequest
	}

	participantsModels := participant_dto.ToManyParticipantModels(dto.Participants)
	transactionDb := s.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			transactionDb.Rollback()
		}
	}()

	err = s.teamRepository.Create(ctx, transactionDb, &teamModel)
	if err != nil {
		transactionDb.Rollback()
		var pqErr *pgconn.PgError
		if errors.As(err, &pqErr) && pqErr.Code == errs.DublicateError {
			log.Error("team already exist", slog.String("team name", teamModel.Name), logger.Err(err))
			return errs.ErrTeamExist.Wrap(fmt.Sprintf("Команда: %s - уже существует", teamModel.Name))
		}
		log.Error("failed to create team", slog.Any("teamModel", teamModel), logger.Err(err))
		return errs.ErrInternalError
	}

	for _, participant := range participantsModels {
		err = s.participantRepository.Create(ctx, transactionDb, &participant)
		if err != nil {
			transactionDb.Rollback()
			log.Error("failed to create participant", slog.Any("participant", participant), logger.Err(err))
			return errs.ErrInternalError
		}
		teamParticipant := team_participant.TeamParticipant{
			TeamID:        teamModel.ID,
			ParticipantID: participant.ID,
		}
		err := s.teamParticipantRepository.Create(ctx, transactionDb, &teamParticipant)
		if err != nil {
			transactionDb.Rollback()
			log.Error("failed to create team participant", slog.Any("teamParticipant", teamParticipant), logger.Err(err))
			return errs.ErrInternalError
		}
	}
	transactionDb.Commit()
	return nil
}

func (s *TeamService) SwitchTeamStatus(ctx context.Context, id string) error {
	const op = "service.TeamService.SwitchTeamStatus"

	log := s.log.With(slog.String("op", op))
	uid, err := uuid.Parse(id)
	if err != nil {
		log.Error("failed parse id to uuid", slog.String("id", id), logger.Err(err))
		return errs.ErrInvalidId
	}

	teamModel, err := s.teamRepository.GetByID(ctx, s.db, uid)
	if err != nil {
		log.Error("failed get team by id", slog.String("id", id), logger.Err(err))
		return errs.ErrInternalError
	}
	if teamModel.Status == team.ReviewStatus {
		teamModel.Status = team.AcceptStatus
	} else {
		teamModel.Status = team.ReviewStatus
	}
	err = s.teamRepository.UpdateByID(ctx, s.db, teamModel)
	if err != nil {
		log.Error("failed update status team", logger.Err(err))
		return errs.ErrInternalError
	}

	return nil
}
