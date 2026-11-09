package quiz_team

import (
	"main/internal/team/model/team"

	"github.com/google/uuid"
)

const (
	OrderNone   = 0
	OrderRed    = 1
	OrderYellow = 2
	OrderGreen  = 3
)

func OrderAllowedMistakes(order int) int {
	switch order {
	case OrderRed:
		return 0
	case OrderYellow:
		return 1
	case OrderGreen:
		return 2
	default:
		return 0
	}
}

func OrderPoints(order int) int {
	switch order {
	case OrderRed:
		return 5
	case OrderYellow:
		return 3
	case OrderGreen:
		return 1
	default:
		return 0
	}
}

type QuizTeam struct {
	ID               uuid.UUID  `gorm:"type:uuid;default:gen_random_uuid();primaryKey" json:"id"`
	EventID          uuid.UUID  `gorm:"type:uuid;not null;index" json:"event_id"`
	TeamID           uuid.UUID  `gorm:"type:uuid;not null;index" json:"team_id"`
	Team             team.Team  `gorm:"foreignKey:TeamID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"team"`
	AccessCode       string     `gorm:"type:varchar(8);not null" json:"access_code"`
	Round1Score      int        `gorm:"type:int;not null;default:0" json:"round1_score"`
	Round2Score      int        `gorm:"type:int;not null;default:0" json:"round2_score"`
	Round3Score      int        `gorm:"type:int;not null;default:0" json:"round3_score"`
	Round3Order      int        `gorm:"type:int;not null;default:0" json:"round3_order"`
	Round3WrongCount int        `gorm:"type:int;not null;default:0" json:"round3_wrong_count"`
	Round3Eliminated bool       `gorm:"type:boolean;not null;default:false" json:"round3_eliminated"`
}
