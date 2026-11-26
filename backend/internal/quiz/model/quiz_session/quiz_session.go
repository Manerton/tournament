package quiz_session

import "github.com/google/uuid"

type QuizSession struct {
	EventID          uuid.UUID `gorm:"type:uuid;primaryKey" json:"event_id"`
	CurrentRound     int       `gorm:"type:int;not null;default:0" json:"current_round"`
	CurrentQuestion  int       `gorm:"type:int;not null;default:0" json:"current_question"`
	AnswerRevealed   bool      `gorm:"type:boolean;not null;default:false" json:"answer_revealed"`
	RoundActive      bool      `gorm:"type:boolean;not null;default:false" json:"round_active"`
	LeaderboardShown bool      `gorm:"type:boolean;not null;default:false" json:"leaderboard_shown"`
	// RulesView is the rules slide on the projector: "none", "overview", "1", "2", "3".
	RulesView string `gorm:"type:varchar(16);not null;default:none" json:"rules_view"`
	// TiebreakActive marks the tie-break (переигровка) mode; TiebreakQuestion is
	// the reserve-pool question (round 4) currently shown, 0 = none.
	TiebreakActive   bool `gorm:"type:boolean;not null;default:false" json:"tiebreak_active"`
	TiebreakQuestion int  `gorm:"type:int;not null;default:0" json:"tiebreak_question"`
}
