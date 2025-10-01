package team_service

import (
	"context"
	"log/slog"
	participant_dto "main/internal/participant/dto"
	"main/internal/pkg/errs"
	"main/internal/pkg/logger"
	"main/internal/team/dto/team_dto"
	"main/internal/team/team_mapper"

	"gorm.io/gorm"
)

type TeamRepositoryInterface struct {
}

type ParticipantRepositoryInterface struct {
}

type TeamService struct {
	teamRepository        TeamRepositoryInterface
	participantRepository ParticipantRepositoryInterface
	log                   *slog.Logger
	db                    *gorm.DB
}

func New(teamRepository TeamRepositoryInterface, participantRepository ParticipantRepositoryInterface, log *slog.Logger, db *gorm.DB) *TeamService {
	return &TeamService{
		teamRepository:        teamRepository,
		participantRepository: participantRepository,
		log:                   log,
		db:                    db,
	}
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

	// err = s.teamRepository.Create(ctx, transactionDb, &teamModel)
	// if err != nil {
	// 	transactionDb.Rollback()
	// 	log.Error("failed to create team", slog.Any("teamModel", teamModel), logger.Err(err))
	// 	return errs.ErrInternalError
	// }

	// for _, participant := range participantsModels {
	// 	s.participantRepository.Create(ctx, transactionDb, &participant)
	// }

	_ = participantsModels
	_ = teamModel

	return nil
}
