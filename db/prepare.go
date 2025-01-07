package db

import (
	"context"
	"fmt"
	"os"

	"morbo/errors"
	"morbo/log"

	"github.com/jackc/pgx/v5/pgxpool"
)

func Prepare(ctx context.Context) (*DB, error) {
	db := DB{log: log.NewLog("db")}

	if err := db.connect(ctx); err != nil {
		db.log.Error.Println("failed to connect to the database")
		return nil, errors.Error
	}

	if err := db.migrate(ctx); err != nil {
		db.log.Error.Println("failed to migrate the database")
		return nil, errors.Error
	}

	db.StartPeriodicStaleSessionsCleanup(ctx)

	return &db, nil
}

func (db *DB) connect(ctx context.Context) error {
	connParamsEnvs := [5]string{
		"MORBO_DB_USER",
		"MORBO_DB_PASSWORD",
		"MORBO_DB_HOST",
		"MORBO_DB_PORT",
		"MORBO_DB_NAME",
	}
	connParams := [len(connParamsEnvs)]any{
		"morbo",
		"morbo",
		"localhost",
		"5432",
		"morbo",
	}

	for index, env := range connParamsEnvs {
		value := os.Getenv(env)
		if value == "" {
			continue
		}
		connParams[index] = value
	}

	connString := fmt.Sprintf("postgres://%s:%s@%s:%s/%s", connParams[:]...)
	config, err := pgxpool.ParseConfig(connString)
	if err != nil {
		db.log.Error.Println(err)
		db.log.Error.Println("failed to parse the database connection string")
		return errors.Error
	}

	pool, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		db.log.Error.Println(err)
		db.log.Error.Println("failed to create a new database connection pool")
		return errors.Error
	}

	if err := pool.Ping(ctx); err != nil {
		db.log.Error.Println(err)
		db.log.Error.Println("failed to ping the database")
		return errors.Error
	}

	db.Pool = pool
	return nil
}

func (db *DB) getCurrentVersion(ctx context.Context) (int, error) {
	var version int

	row := db.Pool.QueryRow(ctx, "SELECT version FROM schema_version LIMIT 1")
	if err := row.Scan(&version); err != nil {
		db.log.Info.Println("assuming the version of the database schema to be 0")
		return 0, nil
	}

	return version, nil
}

func (db *DB) migrate(ctx context.Context) error {
	currentVersion, err := db.getCurrentVersion(ctx)
	if err != nil {
		db.log.Error.Println("failed to get the current schema version")
		return errors.Error
	}

	db.log.Info.Println("current database schema version:", currentVersion)

	for _, migration := range migrations {
		if migration.version > currentVersion {
			db.log.Info.Printf("applying migration to database schema version %d\n", migration.version)
			if _, err := db.Pool.Exec(ctx, migration.sql); err != nil {
				db.log.Error.Printf("failed to apply migration to database schema version %d", migration.version)
				return errors.Error
			}
		}
	}

	return nil
}

func (db *DB) Close() {
	db.Pool.Close()
}
