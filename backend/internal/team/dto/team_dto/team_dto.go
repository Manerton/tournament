package team_dto

import participant_dto "main/internal/participant/dto"

type TeamCreateDTO struct {
	TeamName          string `json:"team_name" binding:"required"`
	ResponsibleName   string `json:"responsible_name"`
	ResponsibleNumber string `json:"responsible_number"`
	EventID           string `json:"event_id" binding:"required,uuid4"`
	SchoolID          string `json:"school_id" binding:"required,uuid4"`
	InitialScore      int    `json:"initial_score"`
}

type TeamResponseDTO struct {
	ID               string `json:"id"`
	Name             string `json:"name"`
	ResponsibleName  string `json:"responsible_name"`
	ResponsiblePhone string `json:"responsible_phone"`
	EventID          string `json:"event_id"`
	SchoolID         string `json:"school_id"`
	Status           int    `json:"status"`
	InitialScore     int    `json:"initial_score"`
}

type TeamUpdateDTO struct {
	Name             string `json:"name" binding:"required"`
	ResponsibleName  string `json:"responsible_name"`
	ResponsiblePhone string `json:"responsible_phone"`
	Status           int    `json:"status"`
}

type RegisterTeamRequestDTO struct {
	TeamCreateDTO
	Participants []participant_dto.ParticipantCreateDTO `json:"participants" binding:"required,dive,required"`
}
