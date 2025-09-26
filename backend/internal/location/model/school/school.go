package school

import (
	"main/internal/location/model/district"

	"github.com/google/uuid"
)

type School struct {
	ID         uuid.UUID         `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	Name       string            `gorm:"type:varchar(255);not null;uniqueIndex"`
	DistrictID uuid.UUID         `gorm:"type:uuid;not null;index"`
	District   district.District `gorm:"foreignKey:DistrictID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
}
