package quiz_tiebreak_team

import "github.com/google/uuid"

// QuizTiebreakTeam is one team of the tied group selected by the host for a
// tie-break. Won marks a team that secured its place and left the active group.
type QuizTiebreakTeam struct {
	EventID uuid.UUID `gorm:"type:uuid;not null;primaryKey" json:"event_id"`
	TeamID  uuid.UUID `gorm:"type:uuid;not null;primaryKey" json:"team_id"`
	Won     bool      `gorm:"type:boolean;not null;default:false" json:"won"`
}

func (QuizTiebreakTeam) TableName() string {
	return "quiz_tiebreak_teams"
}
