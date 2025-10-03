package game_participant_repository

import (
	"context"
	"fmt"
	"main/internal/common/repository"
	"main/internal/game/model/game_participant"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type GameParticipantRepository struct {
	repository.RepositoryBase[game_participant.GameParticipant]
}

func New() *GameParticipantRepository {
	return &GameParticipantRepository{}
}

func (r *GameParticipantRepository) GetAllByGameID(ctx context.Context, db *gorm.DB, gameID uuid.UUID) ([]game_participant.GameParticipant, error) {
	const op = "repository.GameParticipantRepository.GetAllByGameID"
	var GameParticipants []game_participant.GameParticipant
	err := db.WithContext(ctx).Where(game_participant.GameParticipant{GameID: gameID}, gameID).Find(&GameParticipants).Error
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	return GameParticipants, nil
}

func (r *GameParticipantRepository) GetAllByParticipantID(ctx context.Context, db *gorm.DB, participantID uuid.UUID) ([]game_participant.GameParticipant, error) {
	const op = "repository.GameParticipantRepository.GetAllByParticipantID"
	var GameParticipants []game_participant.GameParticipant
	err := db.WithContext(ctx).Where(game_participant.GameParticipant{ParticipantID: participantID}, participantID).Find(&GameParticipants).Error
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	return GameParticipants, nil
}

func (r *GameParticipantRepository) UpdateScore(ctx context.Context, db *gorm.DB, gameID uuid.UUID, participantID uuid.UUID, score int) error {
	const op = "repository.GameParticipantRepository.UpdateScore"
	err := db.WithContext(ctx).Where(game_participant.GameParticipant{GameID: gameID, ParticipantID: participantID}).UpdateColumn("score", gorm.Expr("score + ?", score)).Error
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	return nil
}
