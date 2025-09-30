package country_service

import (
	"context"
	"fmt"
	location_dto "main/internal/location/dto"
	"main/internal/location/location_mapper"
	"main/internal/location/model/country"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type CountryRepositoryInterface interface {
	Create(context.Context, *gorm.DB, *country.Country) error
	UpdateByID(context.Context, *gorm.DB, country.Country) error
	DeleteByID(context.Context, *gorm.DB, uuid.UUID) error
}

type CoutnryService struct {
	repository CountryRepositoryInterface
	db         *gorm.DB
}

func New(db *gorm.DB, repository CountryRepositoryInterface) *CoutnryService {
	return &CoutnryService{
		db:         db,
		repository: repository,
	}
}

func (s *CoutnryService) Create(ctx context.Context, dto location_dto.CountryDTOCreate) error {
	const op = "service.CountryService.Create"

	model := location_mapper.ToModelCountry(dto)
	err := s.repository.Create(ctx, s.db, &model)
	if err != nil {
		return fmt.Errorf("%s: failed create country: %w", op, err)
	}
	return nil
}

func (s *CoutnryService) UpdateByID(ctx context.Context, id string, dto location_dto.CountryDTOCreate) error {
	const op = "service.CoutnryService.UpdateByID"

	uid, err := uuid.Parse(id)
	if err != nil {
		return fmt.Errorf("%s: failed pares uid: %w", op, err)
	}

	model := location_mapper.ToModelCountry(dto)
	model.ID = uid
	err = s.repository.UpdateByID(ctx, s.db, model)
	if err != nil {
		return fmt.Errorf("%s: failed update country: %w", op, err)
	}
	return nil
}

func (s *CoutnryService) DeleteByID(ctx context.Context, id string) error {
	const op = "service.CoutryService.DeleteByID"

	uid, err := uuid.Parse(id)
	if err != nil {
		return fmt.Errorf("%s: failed pares uid: %w", op, err)
	}

	err = s.repository.DeleteByID(ctx, s.db, uid)
	if err != nil {
		return fmt.Errorf("%s: failed delete coutry: %w", op, err)
	}
	return nil
}
