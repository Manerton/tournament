package quiz_buzz

import (
	"time"

	"github.com/google/uuid"
)

const (
	StatusPending    = 1
	StatusCorrect    = 2
	StatusWrong      = 3
	StatusFalseStart = 4
	// StatusOutpaced: the press was recorded inside the 1-second window,
	// but another team pressed first and gets the attempt.
	StatusOutpaced = 5
)

type QuizBuzz struct {
	ID             uuid.UUID `gorm:"type:uuid;default:gen_random_uuid();primaryKey" json:"id"`
	EventID        uuid.UUID `gorm:"type:uuid;not null;index" json:"event_id"`
	Round          int       `gorm:"type:int;not null" json:"round"`
	QuestionNumber int       `gorm:"type:int;not null" json:"question_number"`
	TeamID         uuid.UUID `gorm:"type:uuid;not null;index" json:"team_id"`
	Seq            int64     `gorm:"type:bigint;not null;default:0" json:"seq"`
	Status         int       `gorm:"type:int;not null;default:1" json:"status"`
	BuzzedAt       time.Time `gorm:"type:timestamptz;not null;default:now()" json:"buzzed_at"`
	// OffsetMs is the press time relative to the start of its buzz window; 0 for the winner.
	OffsetMs int `gorm:"type:int;not null;default:0" json:"offset_ms"`
	// WindowIdx numbers the 1-second windows within a question (0, 1, 2, ...).
	WindowIdx int `gorm:"type:int;not null;default:0" json:"window_idx"`
}

func (QuizBuzz) TableName() string {
	return "quiz_buzzes"
}
