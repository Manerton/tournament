package team_dto

import participant_dto "main/internal/participant/dto"

type TeamCreateDTO struct {
	TeamName          string `json:"team_name" binding:"required"`
	ResponsibleName   string `json:"responsible_name"`
	ResponsibleNumber string `json:"responsible_number"`
	EventID           string `json:"event_id" binding:"required,uuid4"`
	SchoolID          string `json:"school_id" binding:"required,uuid4"`
}

type RegisterTeamRequestDTO struct {
	TeamCreateDTO
	Participants []participant_dto.ParticipantCreateDTO `json:"participants" binding:"required,dive,required"`
}
