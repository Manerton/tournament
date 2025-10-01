package participant_dto

import (
	"main/internal/participant/model/participant"

	"github.com/google/uuid"
)

func ToParticipantModel(dto ParticipantCreateDTO) *participant.Participant {
	return &participant.Participant{
		FIO:      dto.FIO,
		Gender:   dto.Gender,
		SchoolID: dto.SchoolID,
		Class:    dto.Class,
	}
}

func ToParticipantResponseDTO(model *participant.Participant) ParticipantResponseDTO {
	return ParticipantResponseDTO{
		ID:       model.ID,
		FIO:      model.FIO,
		Gender:   model.Gender,
		SchoolID: model.SchoolID,
		Class:    model.Class,
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
	if dto.FIO != nil {
		existing.FIO = *dto.FIO
	}
	// if dto.Birthdate != nil {
	// 	existing.Birthdate = *dto.Birthdate
	// }
	if dto.Gender != nil {
		existing.Gender = *dto.Gender
	}
	if dto.SchoolID != nil {
		existing.SchoolID = *dto.SchoolID
	}
	return existing
}
