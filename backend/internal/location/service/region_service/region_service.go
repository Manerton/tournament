package region_servicego

import (
	"context"
	"fmt"
	location_dto "main/internal/location/dto"
	"main/internal/location/location_mapper"
	"main/internal/location/model/region"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type RegionRepositoryInterface interface {
	Create(context.Context, *gorm.DB, region.Region) error
	Delete(context.Context, *gorm.DB, uuid.UUID) error
}

type RegionService struct {
	db         *gorm.DB
	repository RegionRepositoryInterface
}

func New(db *gorm.DB, repository RegionRepositoryInterface) *RegionService {
	return &RegionService{
		db:         db,
		repository: repository,
	}
}

func (s RegionService) Create(ctx context.Context, dto location_dto.RegionDTO) error {
	const op = "service.RegionService.Create"

	model := location_mapper.ToModelRegion(dto)
	err := s.repository.Create(ctx, s.db, model)
	if err != nil {
		return fmt.Errorf("%s: failed create region: %w", op, err)
	}
	return nil
}

func (s *RegionService) DeleteByID(ctx context.Context, id string) error {
	const op = "service.RegionService.DeleteByID"

	uid, err := uuid.Parse(id)
	if err != nil {
		return fmt.Errorf("%s: failed pares uid: %w", op, err)
	}

	err = s.repository.Delete(ctx, s.db, uid)
	if err != nil {
		return fmt.Errorf("%s: failed delete school: %w", op, err)
	}
	return nil
}
