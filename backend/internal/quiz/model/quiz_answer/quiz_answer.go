package quiz_answer

import "github.com/google/uuid"

const (
	GradeWrong   = 1
	GradePartial = 2
	GradeCorrect = 3
)

func GradePoints(round int, grade int) int {
	if round == 2 {
		switch grade {
		case GradeCorrect:
			return 4
		case GradePartial:
			return 2
		default:
			return 0
		}
	}
	return 0
}

type QuizAnswer struct {
	ID             uuid.UUID `gorm:"type:uuid;default:gen_random_uuid();primaryKey" json:"id"`
	EventID        uuid.UUID `gorm:"type:uuid;not null;index" json:"event_id"`
	Round          int       `gorm:"type:int;not null" json:"round"`
	QuestionNumber int       `gorm:"type:int;not null" json:"question_number"`
	TeamID         uuid.UUID `gorm:"type:uuid;not null;index" json:"team_id"`
	Grade          int       `gorm:"type:int;not null" json:"grade"`
	Points         int       `gorm:"type:int;not null;default:0" json:"points"`
}
