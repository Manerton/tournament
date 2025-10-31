package LiveScoreByTour_Repository

import (
	"context"
	"fmt"
	"main/internal/common/repository"
	"main/internal/leaderboard/model/LiveScoreByTour"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type LiveScoreByTourRepository struct {
	repository.RepositoryBase[LiveScoreByTour.LiveScoreByTour]
}

func New() *LiveScoreByTourRepository {
	return &LiveScoreByTourRepository{}
}

func (r *LiveScoreByTourRepository) GetByEventId(ctx context.Context, db *gorm.DB, eventId uuid.UUID) ([]LiveScoreByTour.LiveScoreByTour, error) {
	const op = "repository.LiveScoreByTourRepository.GetByEventId"
	var ScoreByTour []LiveScoreByTour.LiveScoreByTour

	err := db.WithContext(ctx).Preload("Team").Where(LiveScoreByTour.LiveScoreByTour{EventID: eventId}).Find(&ScoreByTour).Error
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	return ScoreByTour, nil
}

func (r *LiveScoreByTourRepository) GetByEventIdTeamId(ctx context.Context, db *gorm.DB, eventId uuid.UUID, teamId uuid.UUID) ([]LiveScoreByTour.LiveScoreByTour, error) {
	const op = "repository.LiveScoreByTourRepository.GetByEventIdTeamId"
	var ScoreByTour []LiveScoreByTour.LiveScoreByTour

	err := db.WithContext(ctx).Preload("Team").Where(LiveScoreByTour.LiveScoreByTour{EventID: eventId}).Where(LiveScoreByTour.LiveScoreByTour{TeamID: teamId}).Find(&ScoreByTour).Error
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	return ScoreByTour, nil
}

func (r *LiveScoreByTourRepository) GetByEventIdParticipantId(ctx context.Context, db *gorm.DB, eventId uuid.UUID, ParticipantId uuid.UUID) ([]LiveScoreByTour.LiveScoreByTour, error) {
	const op = "repository.LiveScoreByTourRepository.GetByEventIdParticipantId"
	var ScoreByTour []LiveScoreByTour.LiveScoreByTour

	err := db.WithContext(ctx).Preload("Team").Where(LiveScoreByTour.LiveScoreByTour{EventID: eventId}).Where(LiveScoreByTour.LiveScoreByTour{ParticipantID: ParticipantId}).Find(&ScoreByTour).Error
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	return ScoreByTour, nil
}

func (r *LiveScoreByTourRepository) UpdateScore(ctx context.Context, db *gorm.DB, eventID uuid.UUID, participantID uuid.UUID, score int) error {
	const op = "repository.GameParticipantRepository.UpdateScore"
	err := db.WithContext(ctx).Model(LiveScoreByTour.LiveScoreByTour{}).
		Where(LiveScoreByTour.LiveScoreByTour{EventID: eventID, ParticipantID: participantID}).
		UpdateColumn("score", gorm.Expr("score + ?", score)).Error
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	return nil
}
