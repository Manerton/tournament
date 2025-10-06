package main

import (
	"flag"
	"log"
	"os"

	_ "github.com/jackc/pgx/v5/stdlib"
	_ "github.com/lib/pq"
	"github.com/pressly/goose/v3"
)

const (
	UP   = "up"
	DOWN = "down"
)

func main() {

	// postgres://postgres:root@localhost:5432/tournamentDB
	const example = "postgres://user:password@localhost:port/dbname"

	var dbDriver, dbStringConnect, migrationPath string

	flag.StringVar(&dbDriver, "driver", "postgres", "driver for db")
	flag.StringVar(&dbStringConnect, "dsn", "", "string for connect to db")
	flag.StringVar(&migrationPath, "migration-path", "./migrations", "path to folder with migration files")

	flag.Parse()

	if dbStringConnect == "" {
		dbStringConnect = os.Getenv("DSN")
	}

	if migrationPath == "" {
		migrationPath = os.Getenv("MIGRATION_PATH")
	}

	if len(flag.Args()) < 1 {
		log.Fatalf("should be use command: UP or DOWM")
	}

	db, err := goose.OpenDBWithDriver(dbDriver, dbStringConnect)
	if err != nil {
		log.Fatalf("failed to connect db: %v, example dsn=%s", err, example)
	}
	defer db.Close()

	command := flag.Arg(0)
	switch command {
	case UP:
		err = goose.Up(db, migrationPath)
	case DOWN:
		err = goose.Down(db, migrationPath)
	default:
		log.Fatalf("unknow command: %s", command)
	}
	if err != nil {
		log.Fatalf("failed to execute command: %s, %v", command, err)
	}
	log.Printf("command %s execute successfully", command)
}
