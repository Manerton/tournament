package district_repository

import (
	"context"
	"fmt"
	"main/internal/location/model/district"
	"main/internal/location/repository/location_repository"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type DistrictRepository struct {
	location_repository.LocationRepositoryBase[district.District]
}

func New() *DistrictRepository {
	return &DistrictRepository{}
}

func (r *DistrictRepository) GetByRegionID(ctx context.Context, db *gorm.DB, regionID uuid.UUID) ([]district.District, error) {
	const op = "repository.DistrictRepository.GetByRegionID"

	var districts []district.District
	err := db.WithContext(ctx).Where(district.District{RegionID: regionID}).Find(&districts).Error
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	return districts, nil
}
