package participant_repository

import (
	"context"
	"fmt"

	models "main/internal/participant/model/participant"

	"gorm.io/gorm"
)

type ParticipantRepository struct{}

// Создать участника
func (r *ParticipantRepository) Create(ctx context.Context, db *gorm.DB, participant *models.Participant) error {
	const op = "repositories.ParticipantRepository.Create"
	if err := db.WithContext(ctx).Create(participant).Error; err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	return nil
}

// Найти участника по ID
func (r *ParticipantRepository) GetByID(ctx context.Context, db *gorm.DB, id string) (*models.Participant, error) {
	const op = "repositories.ParticipantRepository.GetByID"
	var participant models.Participant
	err := db.WithContext(ctx).Where("id = ?", id).First(&participant).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	return &participant, nil
}

// Получить всех участников
func (r *ParticipantRepository) GetAll(ctx context.Context, db *gorm.DB) ([]models.Participant, error) {
	const op = "repositories.ParticipantRepository.GetAll"
	var participants []models.Participant
	if err := db.WithContext(ctx).Find(&participants).Error; err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	return participants, nil
}

// Получить всех по id школы
func (r *ParticipantRepository) GetAllBySchoolID(ctx context.Context, db *gorm.DB, schoolID string) ([]models.Participant, error) {
	const op = "repositories.ParticipantRepository.GetAllBySchoolID"
	var participants []models.Participant
	err := db.WithContext(ctx).
		Where("school_id = ?", schoolID).
		Find(&participants).
		Error
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	return participants, nil
}

// Обновить участника
func (r *ParticipantRepository) Update(ctx context.Context, db *gorm.DB, participant *models.Participant) error {
	const op = "repositories.ParticipantRepository.Update"
	if err := db.WithContext(ctx).Save(participant).Error; err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	return nil
}

// Удалить участника
func (r *ParticipantRepository) Delete(ctx context.Context, db *gorm.DB, id string) error {
	const op = "repositories.ParticipantRepository.Delete"
	if err := db.WithContext(ctx).Delete(&models.Participant{}, "id = ?", id).Error; err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	return nil
}
