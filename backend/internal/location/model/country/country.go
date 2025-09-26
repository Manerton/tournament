package country

import "github.com/google/uuid"

type Country struct {
	ID   uuid.UUID `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	Name string    `gorm:"type:varchar(255);not null;uniqueIndex"`
}
