package district_service

import (
	"context"
	"fmt"
	location_dto "main/internal/location/dto"
	"main/internal/location/location_mapper"
	"main/internal/location/model/district"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type DisctrictRepositoryInterface interface {
	Create(context.Context, *gorm.DB, *district.District) error
	Delete(context.Context, *gorm.DB, uuid.UUID) error
}

type DistrictService struct {
	db         *gorm.DB
	repository DisctrictRepositoryInterface
}

func New(db *gorm.DB, repository DisctrictRepositoryInterface) *DistrictService {
	return &DistrictService{
		db:         db,
		repository: repository,
	}
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

	err = s.repository.Delete(ctx, s.db, uid)
	if err != nil {
		return fmt.Errorf("%s: failed delete school: %w", op, err)
	}
	return nil
}
