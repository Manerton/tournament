package game_participant_repository

import (
	"main/internal/common/repository"
	"main/internal/game/model/game_participant"
)

type GameParticipantRepository struct {
	repository.RepositoryBase[game_participant.GameParticipant]
}

func New() *GameParticipantRepository {
	return &GameParticipantRepository{}
}
