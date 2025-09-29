package region_respository

import (
	"context"
	"fmt"
	"main/internal/location/model/region"
	"main/internal/location/repository/location_repository"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type RegionRepository struct {
	location_repository.LocationRepositoryBase[region.Region]
}

func New() *RegionRepository {
	return &RegionRepository{}
}

func (r *RegionRepository) GetByCountryID(ctx context.Context, db *gorm.DB, countryID uuid.UUID) ([]region.Region, error) {
	const op = "repository.RegionRepository.GetByCountryID"

	var regions []region.Region
	err := db.WithContext(ctx).Where(region.Region{CountryID: countryID}).Find(&regions).Error
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	return regions, nil
}
