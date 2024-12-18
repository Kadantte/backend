package db

import (
	"context"
	"fmt"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"

	"morbo/errors"
	"morbo/log"
)

func Prepare(ctx context.Context) (*DB, error) {
	db := DB{}

	if err := db.connect(ctx); err != nil {
		log.Error.Println("failed to connect to the database")
		return nil, errors.Error
	}

	if err := db.migrate(ctx); err != nil {
		log.Error.Println("failed to migrate the database")
		return nil, errors.Error
	}

	return &db, nil
}

func (db *DB) connect(ctx context.Context) error {
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
		log.Error.Println(err)
		log.Error.Println("failed to parse the database connection string")
		return errors.Error
	}

	pool, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		log.Error.Println(err)
		log.Error.Println("failed to create a new database connection pool")
		return errors.Error
	}

	if err := pool.Ping(ctx); err != nil {
		log.Error.Println(err)
		log.Error.Println("failed to ping the database")
		return errors.Error
	}

	db.Pool = pool
	return nil
}

func (db *DB) getCurrentVersion(ctx context.Context) (int, error) {
	var version int
	row := db.Pool.QueryRow(ctx, "SELECT version FROM schema_version LIMIT 1")
	if err := row.Scan(&version); err != nil {
		log.Info.Println("assuming the version of the database schema to be 0")
		return 0, nil
	}
	return version, nil
}

func (db *DB) migrate(ctx context.Context) error {
	currentVersion, err := db.getCurrentVersion(ctx)
	if err != nil {
		log.Error.Println("failed to get the current schema version")
		return errors.Error
	}
	log.Info.Println("current database schema version:", currentVersion)

	for _, migration := range migrations {
		if migration.version > currentVersion {
			log.Info.Printf("applying migration to database schema version %d\n", migration.version)
			if _, err := db.Pool.Exec(ctx, migration.sql); err != nil {
				log.Error.Printf("failed to apply migration to database schema version %d", migration.version)
				return errors.Error
			}
		}
	}

	return nil
}

func (db *DB) Close() {
	db.Pool.Close()
}
