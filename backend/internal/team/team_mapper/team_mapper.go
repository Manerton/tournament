package team_mapper

import (
	"main/internal/team/dto/team_dto"
	"main/internal/team/model/team"

	"github.com/google/uuid"
)

func ToTeamModel(dto team_dto.TeamCreateDTO) (team.Team, error) {
	eventId, err := uuid.Parse(dto.EventID)
	if err != nil {
		return team.Team{}, err
	}

	shcoolId, err := uuid.Parse(dto.SchoolID)
	if err != nil {
		return team.Team{}, err
	}

	return team.Team{
		Name:             dto.TeamName,
		EventID:          eventId,
		ResponsibleName:  dto.ResponsibleName,
		ResponsiblePhone: dto.ResponsibleNumber,
		SchoolID:         shcoolId,
		Status:           team.ReviewStatus,
		InitialScore:     dto.InitialScore,
	}, nil
}

func ToTeamDTO(model team.Team) team_dto.TeamResponseDTO {
	return team_dto.TeamResponseDTO{
		ID:               model.ID.String(),
		Name:             model.Name,
		ResponsibleName:  model.ResponsibleName,
		ResponsiblePhone: model.ResponsiblePhone,
		EventID:          model.EventID.String(),
		SchoolID:         model.SchoolID.String(),
		Status:           model.Status,
		InitialScore:     model.InitialScore,
	}
}

func ToTeamsDTO(models []team.Team) []team_dto.TeamResponseDTO {
	dtos := make([]team_dto.TeamResponseDTO, 0, len(models))
	for _, model := range models {
		dtos = append(dtos, ToTeamDTO(model))
	}
	return dtos
}
