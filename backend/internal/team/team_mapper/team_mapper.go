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

	return team.Team{
		Name:             dto.TeamName,
		EventID:          eventId,
		ResponsibleName:  dto.ResponsibleName,
		ResponsiblePhone: dto.ResponsibleName,
	}, nil
}
