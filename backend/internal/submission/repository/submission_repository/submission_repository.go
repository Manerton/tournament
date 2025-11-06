package submission_repository

import (
	"context"
	"fmt"
	"main/internal/common/repository"
	"main/internal/submission/model/submission"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type SubmissionRepository struct {
	repository.RepositoryBase[submission.Submission]
}

func New() *SubmissionRepository {
	return &SubmissionRepository{}
}

func (r *SubmissionRepository) GetByEventID(ctx context.Context, db *gorm.DB, eventId uuid.UUID) ([]submission.Submission, error) {
	const op = "repository.SubmissionRepository.GetByEventID"
	result := []submission.Submission{}
	err := db.WithContext(ctx).Model(submission.Submission{}).Where(submission.Submission{EventID: eventId}).Find(&result).Error
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	return result, nil
}

func (r *SubmissionRepository) GetByEventIDAndCode(ctx context.Context, db *gorm.DB, eventId uuid.UUID, code string) (submission.Submission, error) {
	const op = "repository.SubmissionRepository.GetByEventIDAndCode"
	result := submission.Submission{}
	err := db.WithContext(ctx).Model(submission.Submission{}).Where(submission.Submission{Code: code, EventID: eventId}).First(&result).Error
	if err != nil {
		return submission.Submission{}, fmt.Errorf("%s: %w", op, err)
	}
	return result, nil
}
