package participant_dto

import (
	"main/internal/participant/model/participant"

	"github.com/google/uuid"
)

func ToParticipantModel(dto ParticipantCreateDTO) *participant.Participant {
	return &participant.Participant{
		Firstname:  dto.Firstname,
		Surname:    dto.Surname,
		Patronymic: dto.Patronymic,
		Birthdate:  dto.Birthdate,
		Gender:     dto.Gender,
		SchoolID:   dto.SchoolID,
	}
}

func ToParticipantResponseDTO(model *participant.Participant) ParticipantResponseDTO {
	return ParticipantResponseDTO{
		ID:         model.ID,
		Firstname:  model.Firstname,
		Surname:    model.Surname,
		Patronymic: model.Patronymic,
		Birthdate:  model.Birthdate,
		Gender:     model.Gender,
		SchoolID:   model.SchoolID,
	}
}

func ToParticipantsResponseDTO(models []participant.Participant) []ParticipantResponseDTO {
	dtos := make([]ParticipantResponseDTO, len(models))
	for i, m := range models {
		dtos[i] = ToParticipantResponseDTO(&m)
	}
	return dtos
}

// Конвертируем Update DTO в Participant модель для обновления
func ConvertUpdateDTOToParticipant(id uuid.UUID, dto UpdateParticipantDTO, existing participant.Participant) participant.Participant {
	if dto.Firstname != nil {
		existing.Firstname = *dto.Firstname
	}
	if dto.Surname != nil {
		existing.Surname = *dto.Surname
	}
	if dto.Patronymic != nil {
		existing.Patronymic = *dto.Patronymic
	}
	if dto.Birthdate != nil {
		existing.Birthdate = *dto.Birthdate
	}
	if dto.Gender != nil {
		existing.Gender = *dto.Gender
	}
	if dto.SchoolID != nil {
		existing.SchoolID = *dto.SchoolID
	}
	return existing
}
