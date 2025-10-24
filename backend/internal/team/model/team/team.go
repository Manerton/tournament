package team

import (
	"main/internal/event/model/event"
	"main/internal/location/model/school"

	"github.com/google/uuid"
)

const (
	ReviewStatus = 1
	AcceptStatus = 2
)

type Team struct {
	ID               uuid.UUID     `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	Name             string        `gorm:"type:varchar(255);not null;uniqueIndex"`
	ResponsibleName  string        `gorm:"type:varchar(255);not null"`
	ResponsiblePhone string        `gorm:"type:varchar(20);not null"`
	EventID          uuid.UUID     `gorm:"type:uuid;not null;index"`
	Event            event.Event   `gorm:"foreignKey:EventID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
	SchoolID         uuid.UUID     `gorm:"type:uuid;not null;index"`
	School           school.School `gorm:"foreignKey:SchoolID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
	Status           int           `gorm:"type:int;not null;default:1"`
	InitialScore     int           `gorm:"type:int;default:0;"`
}
