package school_service

import (
	"context"
	"fmt"
	location_dto "main/internal/location/dto"
	"main/internal/location/location_mapper"
	"main/internal/location/model/school"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type SchoolRepositoryInterface interface {
	Create(context.Context, *gorm.DB, *school.School) error
	Delete(context.Context, *gorm.DB, uuid.UUID) error
}

type SchoolService struct {
	db         *gorm.DB
	repository SchoolRepositoryInterface
}

func New(db *gorm.DB, repostiory SchoolRepositoryInterface) *SchoolService {
	return &SchoolService{
		db:         db,
		repository: repostiory,
	}
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

	err = s.repository.Delete(ctx, s.db, uid)
	if err != nil {
		return fmt.Errorf("%s: failed delete school: %w", op, err)
	}
	return nil
}
