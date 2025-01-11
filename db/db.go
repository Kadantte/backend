package db

import (
	"morbo/log"

	"github.com/jackc/pgx/v5/pgxpool"
)

type DB struct {
	Pool       *pgxpool.Pool
	log        log.Log
	migrations []migration
}

func NewDB() *DB {
	return &DB{
		Pool:       nil,
		log:        log.NewLog("db"),
		migrations: newMigrations(),
	}
}
