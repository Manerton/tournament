package participant_dto

import (
	"time"

	"github.com/google/uuid"
)

type ParticipantCreateDTO struct {
	Firstname  string    `json:"firstname" validate:"required"`
	Surname    string    `json:"surname" validate:"required"`
	Patronymic string    `json:"patronymic" validate:"required"`
	Birthdate  time.Time `json:"birthdate" validate:"required"`
	Gender     string    `json:"gender" validate:"required,oneof=male female other"`
	SchoolID   uuid.UUID `json:"school_id" validate:"required"`
}

type ParticipantResponseDTO struct {
	ID         uuid.UUID `json:"id"`
	Firstname  string    `json:"firstname"`
	Surname    string    `json:"surname"`
	Patronymic string    `json:"patronymic"`
	Birthdate  time.Time `json:"birthdate"`
	Gender     string    `json:"gender"`
	SchoolID   uuid.UUID `json:"school_id"`
}

// Все поля опциональные
type UpdateParticipantDTO struct {
	Firstname  *string    `json:"firstname,omitempty"`
	Surname    *string    `json:"surname,omitempty"`
	Patronymic *string    `json:"patronymic,omitempty"`
	Birthdate  *time.Time `json:"birthdate,omitempty"`
	Gender     *string    `json:"gender,omitempty"` // можно валидировать
	SchoolID   *uuid.UUID `json:"school_id,omitempty"`
}
