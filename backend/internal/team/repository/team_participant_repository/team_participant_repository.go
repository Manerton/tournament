package team_participant_repository

import (
	"main/internal/common/repository"
	"main/internal/team/model/team_participant"
)

type TeamParticipantRepository struct {
	repository.RepositoryBase[team_participant.TeamParticipant]
}
