package participant_dto

import (
	"main/internal/participant/model/participant"

	"github.com/google/uuid"
)

func ToParticipantModel(dto ParticipantCreateDTO) participant.Participant {
	return participant.Participant{
		FIO:      dto.FIO,
		Gender:   dto.Gender,
		SchoolID: dto.SchoolID,
		Class:    dto.Class,
	}
}

func ToManyParticipantModels(dtos []ParticipantCreateDTO) []participant.Participant {
	models := make([]participant.Participant, 0, len(dtos))
	for _, dto := range dtos {
		models = append(models, ToParticipantModel(dto))
	}
	return models
}

func ToParticipantResponseDTO(model participant.Participant) ParticipantResponseDTO {
	return ParticipantResponseDTO{
		ID:       model.ID,
		FIO:      model.FIO,
		Gender:   model.Gender,
		SchoolID: model.SchoolID,
		Class:    model.Class,
	}
}

func ToParticipantsResponseDTO(models []participant.Participant) []ParticipantResponseDTO {
	dtos := make([]ParticipantResponseDTO, 0, len(models))
	for _, model := range models {
		dtos = append(dtos, ToParticipantResponseDTO(model))
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
