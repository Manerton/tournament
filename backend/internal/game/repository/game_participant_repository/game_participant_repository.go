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
	var result []game_participant.GameParticipant
	err := db.WithContext(ctx).Where(game_participant.GameParticipant{GameID: gameID}, gameID).Find(&result).Error
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	return result, nil
}

func (r *GameParticipantRepository) GetAllByParticipantID(ctx context.Context, db *gorm.DB, participantID uuid.UUID) ([]game_participant.GameParticipant, error) {
	const op = "repository.GameParticipantRepository.GetAllByParticipantID"
	var result []game_participant.GameParticipant
	err := db.WithContext(ctx).Where(game_participant.GameParticipant{ParticipantID: participantID}, participantID).Find(&result).Error
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	return result, nil
}

func (r *GameParticipantRepository) GetByGameParticipantID(ctx context.Context, db *gorm.DB, gameId uuid.UUID, participantID uuid.UUID) (game_participant.GameParticipant, error) {
	const op = "repository.GameParticipantRepository.GetByGameParticipantID"

	var result game_participant.GameParticipant
	err := db.WithContext(ctx).Where(game_participant.GameParticipant{GameID: gameId, ParticipantID: participantID}).First(&result).Error
	if err != nil {
		return game_participant.GameParticipant{}, fmt.Errorf("%s: %w", op, err)
	}

	return result, nil
}

func (r *GameParticipantRepository) UpdateScore(ctx context.Context, db *gorm.DB, gameID uuid.UUID, participantID uuid.UUID, score int) error {
	const op = "repository.GameParticipantRepository.UpdateScore"
	err := db.WithContext(ctx).Model(game_participant.GameParticipant{}).
		Where(game_participant.GameParticipant{GameID: gameID, ParticipantID: participantID}).
		UpdateColumn("score", gorm.Expr("score + ?", score)).Error
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	return nil
}

func (r *GameParticipantRepository) GetByListParticipantIds(ctx context.Context, db *gorm.DB, participantIds []uuid.UUID) ([]game_participant.GameParticipant, error) {
	const op = "repository.GameparticipantRepoository.GetByListParticipantIds"

	var result []game_participant.GameParticipant
	err := db.WithContext(ctx).Model(game_participant.GameParticipant{}).Where("participant_id IN ?", participantIds).Find(&result).Error
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	return result, nil
}
