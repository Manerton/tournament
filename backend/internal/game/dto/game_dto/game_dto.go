package game_dto

type GameDTOResponse struct {
	ID      string `json:"id"`
	EventID string `json:"event_id" binding:"required,uuid"`
	TeamA   string `json:"team_a" binding:"required,uuid"`
	TeamB   string `json:"team_b" binding:"required,uuid"`
}
