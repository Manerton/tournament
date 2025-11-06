package submission_service

import (
	"context"
	"log/slog"
	"main/internal/pkg/errs"
	"main/internal/pkg/logger"
	"main/internal/submission/dto/submission_dto"
	"main/internal/submission/mapper/submission_mapper"
	"main/internal/submission/model/submission"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type SubmissionRepositoryInterface interface {
	GetByEventID(ctx context.Context, db *gorm.DB, eventId uuid.UUID) ([]submission.Submission, error)
	GetByEventIDAndCode(ctx context.Context, db *gorm.DB, eventId uuid.UUID, code string) (submission.Submission, error)
}

type SubmissionService struct {
	db                   *gorm.DB
	log                  *slog.Logger
	submissionRepository SubmissionRepositoryInterface
}

func New(log *slog.Logger, db *gorm.DB, submissionRepository SubmissionRepositoryInterface) *SubmissionService {
	return &SubmissionService{
		db:                   db,
		log:                  log,
		submissionRepository: submissionRepository,
	}
}

func (s *SubmissionService) GetByEventID(ctx context.Context, eventId string) ([]submission_dto.SubmissionDTOResponse, error) {
	const op = "service.SubmissionService.GetByEventID"
	log := s.log.With(slog.String("op", op))

	uid, err := uuid.Parse(eventId)
	if err != nil {
		log.Error("failed parse event id to uuid", slog.String("eventId", eventId), logger.Err(err))
		return nil, errs.ErrBadRequest
	}

	models, err := s.submissionRepository.GetByEventID(ctx, s.db, uid)
	if err != nil {
		log.Error("failed get submissions by eventId", logger.Err(err))
		return nil, errs.ErrInternalError
	}

	return submission_mapper.ManyToDTO(models), nil
}

func (s *SubmissionService) GetByEventIDAndCode(ctx context.Context, eventId string, code string) (submission_dto.SubmissionDTOResponse, error) {
	const op = "service.SubmissionService.GetByEventIDAndCode"
	log := s.log.With(slog.String("op", op))

	uid, err := uuid.Parse(eventId)
	if err != nil {
		log.Error("failed parse id to uuid", slog.String("eventId", eventId), logger.Err(err))
		return submission_dto.SubmissionDTOResponse{}, errs.ErrInternalError
	}

	model, err := s.submissionRepository.GetByEventIDAndCode(ctx, s.db, uid, code)
	if err != nil {
		log.Error("failed get submission by eventId and code", logger.Err(err))
		return submission_dto.SubmissionDTOResponse{}, errs.ErrInternalError
	}

	return submission_mapper.ToDTO(model), nil
}
