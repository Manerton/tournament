package district

import (
	"main/internal/location/model/region"

	"github.com/google/uuid"
)

type District struct {
	ID       uuid.UUID     `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	Name     string        `gorm:"type:varchar(255);not null;uniqueIndex"`
	RegionID uuid.UUID     `gorm:"type:uuid;not null;index"`
	Region   region.Region `gorm:"foreignKey:RegionID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
}
