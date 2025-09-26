package participant

import (
	"main/internal/location/model/school"
	"time"

	"github.com/google/uuid"
)

type Participant struct {
	ID         uuid.UUID     `gorm:"type:uuid;defaul:gen_random_uuid();primaryKey"`
	Firstname  string        `gorm:"type:varchar(255);not null"`
	Surname    string        `gorm:"type:varchar(255);not null"`
	Patronymic string        `gorm:"type:varchar(255);not null"`
	Birthdate  time.Time     `gorm:"type:date;not null"`
	Gender     string        `gorm:"type:varchar(50);not null"` // TODO: Define enum type
	SchoolID   uuid.UUID     `gorm:"type:uuid;not null;index"`
	School     school.School `gorm:"foreignKey:SchoolID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
}
