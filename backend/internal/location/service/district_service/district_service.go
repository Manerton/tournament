package district_service

import (
	"context"
	"fmt"
	"log/slog"
	location_dto "main/internal/location/dto"
	"main/internal/location/location_mapper"
	"main/internal/location/model/district"
	"main/internal/pkg/errs"
	"main/internal/pkg/logger"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type DisctrictRepositoryInterface interface {
	GetAll(context.Context, *gorm.DB) ([]district.District, error)

	Create(context.Context, *gorm.DB, *district.District) error
	DeleteByID(context.Context, *gorm.DB, uuid.UUID) error
}

type DistrictService struct {
	log        *slog.Logger
	db         *gorm.DB
	repository DisctrictRepositoryInterface
}

func New(log *slog.Logger, db *gorm.DB, repository DisctrictRepositoryInterface) *DistrictService {
	return &DistrictService{
		log:        log,
		db:         db,
		repository: repository,
	}
}

func (s DistrictService) GetAll(ctx context.Context) ([]location_dto.DistrictDTOResponse, error) {
	const op = "service.DistrictService.GetAll"

	log := s.log.With(slog.String("op", op))

	models, err := s.repository.GetAll(ctx, s.db)
	if err != nil {
		log.Error("failed get all districts", logger.Err(err))
		return nil, errs.ErrInternalError
	}

	return location_mapper.ToManyDTODistrict(models), nil
}

func (s DistrictService) Create(ctx context.Context, dto location_dto.DistrictDTO) error {
	const op = "service.DistrictService.Create"

	model := location_mapper.ToModelDistrict(dto)
	err := s.repository.Create(ctx, s.db, &model)
	if err != nil {
		return fmt.Errorf("%s: failed create region: %w", op, err)
	}
	return nil
}

func (s *DistrictService) DeleteByID(ctx context.Context, id string) error {
	const op = "service.DistrictService.DeleteByID"

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
