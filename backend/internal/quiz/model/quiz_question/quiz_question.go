package quiz_question

import "github.com/google/uuid"

type QuizQuestion struct {
	ID             uuid.UUID `gorm:"type:uuid;default:gen_random_uuid();primaryKey" json:"id"`
	EventID        uuid.UUID `gorm:"type:uuid;not null;index" json:"event_id"`
	Round          int       `gorm:"type:int;not null;index" json:"round"`
	QuestionNumber int       `gorm:"type:int;not null" json:"question_number"`
	Text           string    `gorm:"type:text;not null" json:"text"`
	Answer         string    `gorm:"type:text;not null" json:"answer"`
	Explanation    string    `gorm:"type:text;not null" json:"explanation"`
	Criteria       string    `gorm:"type:text;not null" json:"criteria"`
	// Used marks a reserve-pool (round 4) question as shown: the hall has seen it,
	// so it must never be shown again.
	Used bool `gorm:"type:boolean;not null;default:false" json:"used"`
}
