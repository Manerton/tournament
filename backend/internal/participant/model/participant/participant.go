package participant

import (
	"main/internal/location/model/school"
	//"time"

	"github.com/google/uuid"
)

type Participant struct {
	ID  uuid.UUID `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	FIO string    `gorm:"type:varchar(255);not null"`
	// Birthdate  time.Time     `gorm:"type:date;not null"`
	Class        int           `gorm:"type:smallint;not null"`
	Gender       int           `gorm:"type:smallint;not null"` // или "type:tinyint"
	SchoolID     uuid.UUID     `gorm:"type:uuid;not null;index"`
	School       school.School `gorm:"foreignKey:SchoolID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
	InitialScore int           `gorm:"type:int;default 0;"`
}
