package quiz_session

import "github.com/google/uuid"

type QuizSession struct {
	EventID          uuid.UUID `gorm:"type:uuid;primaryKey" json:"event_id"`
	CurrentRound     int       `gorm:"type:int;not null;default:0" json:"current_round"`
	CurrentQuestion  int       `gorm:"type:int;not null;default:0" json:"current_question"`
	AnswerRevealed   bool      `gorm:"type:boolean;not null;default:false" json:"answer_revealed"`
	RoundActive      bool      `gorm:"type:boolean;not null;default:false" json:"round_active"`
	LeaderboardShown bool      `gorm:"type:boolean;not null;default:false" json:"leaderboard_shown"`
}
