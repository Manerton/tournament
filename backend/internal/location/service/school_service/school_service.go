package school_service

import (
	"context"
	"fmt"
	"log/slog"
	location_dto "main/internal/location/dto"
	"main/internal/location/location_mapper"
	"main/internal/location/model/school"
	"main/internal/pkg/errs"
	"main/internal/pkg/logger"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type SchoolRepositoryInterface interface {
	GetAll(context.Context, *gorm.DB) ([]school.School, error)
	GetByDistrictID(context.Context, *gorm.DB, uuid.UUID) ([]school.School, error)

	Create(context.Context, *gorm.DB, *school.School) error
	DeleteByID(context.Context, *gorm.DB, uuid.UUID) error
}

type SchoolService struct {
	log        *slog.Logger
	db         *gorm.DB
	repository SchoolRepositoryInterface
}

func New(log *slog.Logger, db *gorm.DB, repostiory SchoolRepositoryInterface) *SchoolService {
	return &SchoolService{
		log:        log,
		db:         db,
		repository: repostiory,
	}
}

func (s SchoolService) GetByDistrictId(ctx context.Context, id string) ([]location_dto.SchoolDTOResponse, error) {
	const op = "service.SchoolService.GetByDistrictId"

	log := s.log.With(slog.String("op", op))

	uid, err := uuid.Parse(id)
	if err != nil {
		log.Error("failed parse id to uuid", slog.String("id", id), logger.Err(err))
		return nil, errs.ErrInvalidId
	}

	models, err := s.repository.GetByDistrictID(ctx, s.db, uid)
	if err != nil {
		log.Error("failed get all schools", logger.Err(err))
		return nil, errs.ErrInternalError
	}

	return location_mapper.ToManyDTOSchools(models), nil
}

func (s *SchoolService) Create(ctx context.Context, dto location_dto.SchoolDTO) error {
	const op = "service.SchoolService.Create"

	model := location_mapper.ToModelSchool(dto)
	err := s.repository.Create(ctx, s.db, &model)
	if err != nil {
		return fmt.Errorf("%s: failed create school: %w", op, err)
	}
	return nil
}

func (s *SchoolService) DeleteByID(ctx context.Context, id string) error {
	const op = "service.SchoolService.DeleteByID"

	uid, err := uuid.Parse(id)
	if err != nil {
		return fmt.Errorf("%s: failed pares uid: %w", op, err)
	}

	err = s.repository.DeleteByID(ctx, s.db, uid)
	if err != nil {
		return fmt.Errorf("%s: failed delete school: %w", op, err)
	}
	return nil
}
