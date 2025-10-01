package participant_dto

import (
	//"time"

	"github.com/google/uuid"
)

type ParticipantCreateDTO struct {
	FIO string `json:"FIO" validate:"required"`
	//Birthdate  time.Time `json:"birthdate" validate:"required"`
	Gender   int       `json:"gender" validate:"required"`
	Class    int       `json:"class" validate:"required"`
	SchoolID uuid.UUID `json:"school_id" validate:"required"`
}

type ParticipantResponseDTO struct {
	ID       uuid.UUID `json:"id"`
	FIO      string    `json:"FIO"`
	Gender   int       `json:"gender"`
	Class    int       `json:"class" validate:"required"`
	SchoolID uuid.UUID `json:"school_id"`
}

// Все поля опциональные
type UpdateParticipantDTO struct {
	FIO      *string    `json:"FIO,omitempty"`
	Gender   *int       `json:"gender,omitempty"`
	Class    *int       `json:"class,omitempty"`
	SchoolID *uuid.UUID `json:"school_id,omitempty"`
}
