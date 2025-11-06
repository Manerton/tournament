package submission

import "github.com/google/uuid"

type Submission struct {
	ID      uuid.UUID `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	EventID uuid.UUID `gorm:"type:uuid;not null;uniqueIndex:idx_code_event"`
	Code    string    `gorm:"type:varchar(255);not null;"`
	Answer  string    `gorm:"type:varchar(255);not null;"`
}
