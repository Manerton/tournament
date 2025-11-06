package submission_mapper

import (
	"main/internal/submission/dto/submission_dto"
	"main/internal/submission/model/submission"
)

func ToDTO(model submission.Submission) submission_dto.SubmissionDTOResponse {
	return submission_dto.SubmissionDTOResponse{
		ID:      model.ID.String(),
		EventId: model.EventID.String(),
		Code:    model.Code,
		Answer:  model.Answer,
	}
}

func ManyToDTO(models []submission.Submission) []submission_dto.SubmissionDTOResponse {
	result := make([]submission_dto.SubmissionDTOResponse, 0, len(models))
	for _, model := range models {
		result = append(result, ToDTO(model))
	}
	return result
}
