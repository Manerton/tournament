package user

import (
	"time"

	"github.com/google/uuid"
)

type RoleType int

const (
	ADMIN = 1
	JUDGE = 2
)

/*func (r RoleType) String() string {
	switch r {
	case ADMIN:
		return "ADMIN"
	case JUDGE:
		return "JUDGE"
	default:
		return "UNKNOWN"
	}
}

// Чтобы при отдаче в JSON было "ADMIN", а не 1
func (r RoleType) MarshalJSON() ([]byte, error) {
	return json.Marshal(r.String())
}*/

type User struct {
	ID        uuid.UUID `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	Login     string    `gorm:"type:varchar(255);not null;uniqueIndex"`
	Password  string    `gorm:"type:varchar(255);not null"`
	Role      RoleType
	CreatedAt time.Time `gorm:"not null;default:now()"`
	UpdatedAt time.Time `gorm:"not null;default:now()"`
}
