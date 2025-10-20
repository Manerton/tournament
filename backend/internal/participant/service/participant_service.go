package participant_service

import (
	"context"
	"fmt"
	"log/slog"

	participant_dto "main/internal/participant/dto"
	"main/internal/participant/model/participant"
	"main/internal/pkg/errs"
	"main/internal/pkg/logger"
	"main/internal/team/model/team_participant"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type ParticipantRepositoryInterface interface {
	Create(ctx context.Context, db *gorm.DB, participant *participant.Participant) error
	GetByID(ctx context.Context, db *gorm.DB, id uuid.UUID) (participant.Participant, error)
	GetAll(ctx context.Context, db *gorm.DB) ([]participant.Participant, error)
	GetAllBySchoolID(ctx context.Context, db *gorm.DB, schoolID uuid.UUID) ([]participant.Participant, error)
	UpdateByID(ctx context.Context, db *gorm.DB, participant participant.Participant) error
	DeleteByID(ctx context.Context, db *gorm.DB, id uuid.UUID) error
}

type TeamParticipantRepositoryInterface interface {
	GetAllByTeamID(ctx context.Context, db *gorm.DB, teamID uuid.UUID) ([]team_participant.TeamParticipant, error)
}

// сервис
type ParticipantService struct {
	db                        *gorm.DB
	participantRepository     ParticipantRepositoryInterface
	teamParticipantRepository TeamParticipantRepositoryInterface
	log                       *slog.Logger
}

func New(log *slog.Logger, db *gorm.DB, repo ParticipantRepositoryInterface, teamRepo TeamParticipantRepositoryInterface) *ParticipantService {
	return &ParticipantService{
		log:                       log,
		db:                        db,
		participantRepository:     repo,
		teamParticipantRepository: teamRepo,
	}
}

// Create
func (s *ParticipantService) Create(ctx context.Context, createDTO participant_dto.ParticipantCreateDTO) (uuid.UUID, error) {
	const op = "services.participant_service.Create"

	entityPtr := participant_dto.ToParticipantModel(createDTO)
	err := s.participantRepository.Create(ctx, s.db, &entityPtr)
	if err != nil {
		return uuid.Nil, fmt.Errorf("%s: %w", op, err)
	}
	return entityPtr.ID, nil
}

// GetByID
func (s *ParticipantService) GetByID(ctx context.Context, id string) (participant_dto.ParticipantResponseDTO, error) {
	const op = "services.participant_service.GetByID"

	uid, err := uuid.Parse(id)
	if err != nil {
		return participant_dto.ParticipantResponseDTO{}, fmt.Errorf("%s: invalid uuid: %w", op, err)
	}

	entity, err := s.participantRepository.GetByID(ctx, s.db, uid)
	if err != nil {
		return participant_dto.ParticipantResponseDTO{}, fmt.Errorf("%s: %w", op, err)
	}

	return participant_dto.ToParticipantResponseDTO(entity), nil
}

// GetAll
func (s *ParticipantService) GetAll(ctx context.Context) ([]participant_dto.ParticipantResponseDTO, error) {
	const op = "services.participant_service.GetAll"

	entities, err := s.participantRepository.GetAll(ctx, s.db)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return participant_dto.ToParticipantsResponseDTO(entities), nil
}

// GetAllBySchoolID
func (s *ParticipantService) GetAllBySchoolID(ctx context.Context, schoolID string) ([]participant_dto.ParticipantResponseDTO, error) {
	const op = "services.participant_service.GetAllBySchoolID"

	uid, err := uuid.Parse(schoolID)
	if err != nil {
		return nil, fmt.Errorf("%s: invalid uuid: %w", op, err)
	}

	entities, err := s.participantRepository.GetAllBySchoolID(ctx, s.db, uid)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return participant_dto.ToParticipantsResponseDTO(entities), nil
}

func (s *ParticipantService) GetAllByTeamID(ctx context.Context, teamId string) ([]participant_dto.ParticipantResponseDTO, error) {
	const op = "service.ParticipantService.GetAllByTeamID"
	log := s.log.With(slog.String("op", op))
	uid, err := uuid.Parse(teamId)
	if err != nil {
		log.Error("failed parse id to uuid", slog.String("id", teamId), logger.Err(err))
		return nil, errs.ErrInvalidId
	}

	teamParticipantModels, err := s.teamParticipantRepository.GetAllByTeamID(ctx, s.db, uid)
	if err != nil {
		log.Error("failed get team participants by team id", logger.Err(err))
		return nil, errs.ErrInternalError
	}
	participantModels := make([]participant.Participant, 0, len(teamParticipantModels))
	for _, tp := range teamParticipantModels {
		participantModel, err := s.participantRepository.GetByID(ctx, s.db, tp.ParticipantID)
		if err != nil {
			log.Error("failed get participant by id", logger.Err(err))
			return nil, errs.ErrInternalError
		}
		participantModels = append(participantModels, participantModel)
	}

	return participant_dto.ToParticipantsResponseDTO(participantModels), nil
}

func (s *ParticipantService) Update(ctx context.Context, id string, updateDTO participant_dto.UpdateParticipantDTO) error {
	const op = "services.participant_service.Update"

	uid, err := uuid.Parse(id)
	if err != nil {
		return fmt.Errorf("%s: invalid uuid: %w", op, err)
	}

	// Получаем текущего участника
	existing, err := s.participantRepository.GetByID(ctx, s.db, uid)
	if err != nil {
		return fmt.Errorf("%s: failed to get participant: %w", op, err)
	}
	if existing.ID == uuid.Nil {
		return fmt.Errorf("%s: participant not found", op)
	}

	// Применяем изменения из DTO
	updated := participant_dto.ConvertUpdateDTOToParticipant(uid, updateDTO, existing)

	// Сохраняем обновленного участника
	if err := s.participantRepository.UpdateByID(ctx, s.db, updated); err != nil {
		return fmt.Errorf("%s: failed to update participant: %w", op, err)
	}

	return nil
}

// Delete
func (s *ParticipantService) Delete(ctx context.Context, id string) error {
	const op = "services.participant_service.Delete"

	uid, err := uuid.Parse(id)
	if err != nil {
		return fmt.Errorf("%s: invalid uuid: %w", op, err)
	}

	if err := s.participantRepository.DeleteByID(ctx, s.db, uid); err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	return nil
}
