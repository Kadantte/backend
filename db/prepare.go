// Copyright (C) 2024 Pavel Sobolev
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published
// by the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU Affero General Public License for more details.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program.  If not, see <https://www.gnu.org/licenses/>.

package db

import (
	"context"
	"fmt"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"

	"morbo/errors"
	"morbo/log"
)

func Prepare() (*DB, error) {
	db := DB{}

	if err := db.connect(); err != nil {
		log.Error.Println("failed to connect to the database")
		return nil, errors.Error
	}

	if err := db.migrate(); err != nil {
		log.Error.Println("failed to migrate the database")
		return nil, errors.Error
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

	db.pool = pool
	return nil
}

func (db *DB) getCurrentVersion() (int, error) {
	var version int
	row := db.pool.QueryRow(context.Background(), "SELECT version FROM schema_version LIMIT 1")
	if err := row.Scan(&version); err != nil {
		log.Info.Println("assuming the version of the database schema to be 0")
		return 0, nil
	}
	return version, nil
}

func (db *DB) migrate() error {
	ctx := context.Background()

	currentVersion, err := db.getCurrentVersion()
	if err != nil {
		log.Error.Println("failed to get the current schema version")
		return errors.Error
	}
	log.Info.Println("current database schema version:", currentVersion)

	for _, migration := range migrations {
		if migration.version > currentVersion {
			log.Info.Printf("applying migration to database schema version %d\n", migration.version)
			if _, err := db.pool.Exec(ctx, migration.sql); err != nil {
				log.Error.Printf("failed to apply migration to database schema version %d", migration.version)
				return errors.Error
			}
		}
	}

	return nil
}

func (db *DB) Close() {
	db.pool.Close()
}
