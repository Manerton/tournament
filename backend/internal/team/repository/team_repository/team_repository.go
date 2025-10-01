package team_repository

import (
	"main/internal/common/repository"
	"main/internal/team/model/team"
)

type TeamRepository struct {
	repository.RepositoryBase[team.Team]
}
