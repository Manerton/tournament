package participant_repository

import (
	"context"
	"fmt"

	"main/internal/common/repository"
	"main/internal/participant/model/participant"

	"gorm.io/gorm"
)

type ParticipantRepository struct {
	repository.RepositoryBase[participant.Participant]
}

func New() *ParticipantRepository {
	return &ParticipantRepository{}
}

// Получить всех по id школы
func (r *ParticipantRepository) GetAllBySchoolID(ctx context.Context, db *gorm.DB, schoolID string) ([]participant.Participant, error) {
	const op = "repositories.ParticipantRepository.GetAllBySchoolID"
	var participants []participant.Participant
	err := db.WithContext(ctx).
		Where("school_id = ?", schoolID).
		Find(&participants).
		Error
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	return participants, nil
}
