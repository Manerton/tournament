package postgresql

import (
	"fmt"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func MustPostgreSQL(connectStr string) *gorm.DB {
	const op = "storage.postgresql.MustPostgreSQL"
	db, err := gorm.Open(postgres.Open(connectStr), &gorm.Config{})
	if err != nil {
		panic(fmt.Errorf("%s: failed to connect to PostgreSQL: %w", op, err))
	}

	return db
}
