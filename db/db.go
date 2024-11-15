package db

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"

	"morbo/errors"
)

type DB struct {
	pool *pgxpool.Pool
}

func Prepare() (*DB, error) {
	db := DB{}

	if err := db.connect(); err != nil {
		return nil, errors.Chain("failed to connect to the database", err)
	}

	if err := db.migrate(); err != nil {
		return nil, errors.Chain("failed to migrate the database", err)
	}

	return &db, nil
}

func (db *DB) connect() error {
	ctx := context.Background()

	dbHost := os.Getenv("MORBO_DB_HOST")
	if dbHost == "" {
		dbHost = "localhost"
	}
	dbPort := os.Getenv("MORBO_DB_PORT")
	if dbPort == "" {
		dbPort = "5432"
	}
	dbUser := os.Getenv("MORBO_DB_USER")
	if dbUser == "" {
		dbUser = "morbo"
	}
	dbPassword := os.Getenv("MORBO_DB_PASSWORD")
	if dbPassword == "" {
		dbPassword = "morbo"
	}
	dbName := os.Getenv("MORBO_DB_NAME")
	if dbName == "" {
		dbName = "morbo"
	}

	connString := fmt.Sprintf("postgres://%s:%s@%s:%s/%s", dbUser, dbPassword, dbHost, dbPort, dbName)
	config, err := pgxpool.ParseConfig(connString)
	if err != nil {
		return err
	}

	pool, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		return err
	}

	if err := pool.Ping(ctx); err != nil {
		return err
	}

	db.pool = pool
	return nil
}

func (db *DB) getCurrentVersion() (int, error) {
	var version int
	row := db.pool.QueryRow(context.Background(), "SELECT version FROM schema_version LIMIT 1")
	if err := row.Scan(&version); err != nil {
		return 0, nil
	}
	return version, nil
}

func (db *DB) migrate() error {
	ctx := context.Background()

	currentVersion, err := db.getCurrentVersion()
	if err != nil {
		return errors.Chain("failed to get the current schema version", err)
	}
	log.Println("current schema version:", currentVersion)

	for _, migration := range migrations {
		if migration.version > currentVersion {
			fmt.Printf("applying migration to schema version %d\n", migration.version)
			if _, err := db.pool.Exec(ctx, migration.sql); err != nil {
				log.Printf("failed to apply migration to schema version %d: %v", migration.version, err)
				return err
			}
		}
	}

	return nil
}

func (db *DB) Close() {
	db.pool.Close()
}
