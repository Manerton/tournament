package leaderboard_dto

type LeaderboardDTOResponse struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Score int    `json:"score"`
}

type LeaderboardPlayesDTOResponse struct {
	TeamName string `json:"team_name"`
	Name     string `json:"name"`
	Score    int    `json:"score"`
}
