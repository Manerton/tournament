package game_dto

type GameDTOResponse struct {
	ID      string `json:"id"`
	EventID string `json:"event_id" binding:"required,uuid"`
	TeamAId string `json:"team_a" binding:"required,uuid"`
	TeamBId string `json:"team_b" binding:"required,uuid"`
}

type GameAllInfoResponse struct {
	ID         string `json:"id"`
	TeamAId    string `json:"team_a_id"`
	TeamAScore int    `json:"team_a_score"`
	TeamAName  string `json:"team_a_name"`
	TeamBId    string `json:"team_b_id"`
	TeamBScore int    `json:"team_b_score"`
	TeamBName  string `json:"team_b_name"`
	GameStatus int    `json:"game_status"`
}
