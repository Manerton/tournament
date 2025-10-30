package game_mapper

import (
	"main/internal/game/dto/game_dto"
	"main/internal/game/model/game"
	"main/internal/game/model/game_participant"
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

func ToDTOGameParticipant(gameParticipant game_participant.GameParticipant) game_dto.GameParticipantDTOResponse {
	return game_dto.GameParticipantDTOResponse{
		ID:            gameParticipant.ID.String(),
		GameID:        gameParticipant.GameID.String(),
		ParticipantID: gameParticipant.ParticipantID.String(),
		Score:         gameParticipant.Score,
	}
}

func ToManyDTOGameParticipants(gameParticipants []game_participant.GameParticipant) []game_dto.GameParticipantDTOResponse {
	gameParticipantsDTO := make([]game_dto.GameParticipantDTOResponse, 0, len(gameParticipants))
	for _, gp := range gameParticipants {
		gameParticipantsDTO = append(gameParticipantsDTO, ToDTOGameParticipant(gp))
	}
	return gameParticipantsDTO
}
