package school_repository

import (
	"context"
	"fmt"
	"main/internal/common/repository"
	"main/internal/location/model/school"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type SchoolRepository struct {
	repository.RepositoryBase[school.School]
}

func New() *SchoolRepository {
	return &SchoolRepository{}
}

func (r *SchoolRepository) GetByDistrictID(ctx context.Context, db *gorm.DB, districtID uuid.UUID) ([]school.School, error) {
	const op = "repository.SchoolRepository.GetByDistrictID"

	var schools []school.School
	err := db.WithContext(ctx).Where(school.School{DistrictID: districtID}).Find(&schools).Error
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	return schools, nil
}
