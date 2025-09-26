package region

import (
	"main/internal/location/model/country"

	"github.com/google/uuid"
)

type Region struct {
	ID        uuid.UUID       `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	Name      string          `gorm:"type:varchar(255);not null;uniqueIndex"`
	CountryID uuid.UUID       `gorm:"type:uuid;not null;index"`
	Country   country.Country `gorm:"foreignKey:CountryID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
}
