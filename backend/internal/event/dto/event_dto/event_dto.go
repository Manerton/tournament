package event_dto

type EventDTOCreate struct {
	Name            string  `json:"name" binding:"required"`
	StartDate       string  `json:"start_date" binding:"required,datetime=2006-01-02"`
	EndDate         string  `json:"end_date" binding:"required,datetime=2006-01-02"`
	PreviousEventID *string `json:"previous_event_id" binding:"omitempty,uuid"`
	TypeDraws       int     `json:"type_draws" binding:"required,oneof=1 2 3"` // 1: Knockout, 2: RoundRobin, 3: Swiss
}

type EventDTOUpdate struct {
	Name      *string `json:"name"`
	StartDate *string `json:"start_date"`
	EndDate   *string `json:"end_date"`
	Status    *int    `json:"status"`
}

type EventDTOResponse struct {
	ID              string `json:"id"`
	Name            string `json:"name" binding:"required"`
	StartDate       string `json:"start_date" binding:"required,datetime=2006-01-02"`
	EndDate         string `json:"end_date" binding:"required,datetime=2006-01-02"`
	PreviousEventID string `json:"previous_event_id" binding:"omitempty,uuid"`
	Status          int    `json:"status"`
	Type            int    `json:"type"`
	TypeDraws       int    `json:"type_draws"`
	Stage           int    `json:"stage"`
	MaxStage        int    `json:"max_stage"`
}
