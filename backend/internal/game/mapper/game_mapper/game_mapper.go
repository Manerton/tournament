package game_mapper

import (
	"main/internal/game/dto/game_dto"
	"main/internal/game/model/game"
)

func ToDTOGame(game game.Game) game_dto.GameDTOResponse {
	return game_dto.GameDTOResponse{
		ID:      game.ID.String(),
		EventID: game.EventID.String(),
		TeamAId: game.TeamAId.String(),
		TeamBId: game.TeamBId.String(),
	}
}

func ToManyDTOGame(games []game.Game) []game_dto.GameDTOResponse {
	gamesDTO := make([]game_dto.GameDTOResponse, 0, len(games))
	for _, g := range games {
		gamesDTO = append(gamesDTO, ToDTOGame(g))
	}
	return gamesDTO
}
