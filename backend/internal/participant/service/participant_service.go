package participant_service

import (
	"context"
	"fmt"

	dto "main/internal/participant/dto"
	model "main/internal/participant/model/participant"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type ParticipantRepository interface {
	Create(ctx context.Context, db *gorm.DB, participant model.Participant) (uuid.UUID, error)
	GetByID(ctx context.Context, db *gorm.DB, id uuid.UUID) (model.Participant, error)
	GetAll(ctx context.Context, db *gorm.DB) ([]model.Participant, error)
	GetAllBySchoolID(ctx context.Context, db *gorm.DB, schoolID uuid.UUID) ([]model.Participant, error)
	Update(ctx context.Context, db *gorm.DB, participant model.Participant) error
	Delete(ctx context.Context, db *gorm.DB, id uuid.UUID) error
}

// сервис
type ParticipantService struct {
	db         *gorm.DB
	repository ParticipantRepository
}

func NewParticipantService(db *gorm.DB, repo ParticipantRepository) *ParticipantService {
	return &ParticipantService{
		db:         db,
		repository: repo,
	}
}

// Create
func (s *ParticipantService) Create(ctx context.Context, createDTO dto.ParticipantCreateDTO) (uuid.UUID, error) {
	const op = "services.participant_service.Create"

	entityPtr := dto.ToParticipantModel(createDTO)
	id, err := s.repository.Create(ctx, s.db, *entityPtr)
	if err != nil {
		return uuid.Nil, fmt.Errorf("%s: %w", op, err)
	}
	return id, nil
}

// GetByID
func (s *ParticipantService) GetByID(ctx context.Context, id string) (dto.ParticipantResponseDTO, error) {
	const op = "services.participant_service.GetByID"

	uid, err := uuid.Parse(id)
	if err != nil {
		return dto.ParticipantResponseDTO{}, fmt.Errorf("%s: invalid uuid: %w", op, err)
	}

	entity, err := s.repository.GetByID(ctx, s.db, uid)
	if err != nil {
		return dto.ParticipantResponseDTO{}, fmt.Errorf("%s: %w", op, err)
	}

	return dto.ToParticipantResponseDTO(&entity), nil
}

// GetAll
func (s *ParticipantService) GetAll(ctx context.Context) ([]dto.ParticipantResponseDTO, error) {
	const op = "services.participant_service.GetAll"

	entities, err := s.repository.GetAll(ctx, s.db)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return dto.ToParticipantsResponseDTO(entities), nil
}

// GetAllBySchoolID
func (s *ParticipantService) GetAllBySchoolID(ctx context.Context, schoolID string) ([]dto.ParticipantResponseDTO, error) {
	const op = "services.participant_service.GetAllBySchoolID"

	uid, err := uuid.Parse(schoolID)
	if err != nil {
		return nil, fmt.Errorf("%s: invalid uuid: %w", op, err)
	}

	entities, err := s.repository.GetAllBySchoolID(ctx, s.db, uid)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return dto.ToParticipantsResponseDTO(entities), nil
}

func (s *ParticipantService) Update(ctx context.Context, id string, updateDTO dto.UpdateParticipantDTO) error {
	const op = "services.participant_service.Update"

	uid, err := uuid.Parse(id)
	if err != nil {
		return fmt.Errorf("%s: invalid uuid: %w", op, err)
	}

	// Получаем текущего участника
	existing, err := s.repository.GetByID(ctx, s.db, uid)
	if err != nil {
		return fmt.Errorf("%s: failed to get participant: %w", op, err)
	}
	if existing.ID == uuid.Nil {
		return fmt.Errorf("%s: participant not found", op)
	}

	// Применяем изменения из DTO
	updated := dto.ConvertUpdateDTOToParticipant(uid, updateDTO, existing)

	// Сохраняем обновленного участника
	if err := s.repository.Update(ctx, s.db, updated); err != nil {
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

	if err := s.repository.Delete(ctx, s.db, uid); err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	return nil
}
