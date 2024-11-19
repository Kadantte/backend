package db

type migration struct {
	version int
	sql     string
}

var migrations = []migration{
	{
		version: 1,
		sql: `
			CREATE TABLE schema_version (
				version INT PRIMARY KEY
			);

			INSERT INTO schema_version (version) VALUES (1);
	    `,
	},
	{
		version: 2,
		sql: `
			CREATE TABLE users(
				id INTEGER NOT NULL PRIMARY KEY,
				username TEXT NOT NULL UNIQUE,
				password TEXT NOT NULL
			);

			CREATE TABLE sessions(
				session_token TEXT NOT NULL PRIMARY KEY,
				user_id INTEGER NOT NULL
			);

			UPDATE schema_version SET version = 2;
		`,
	},
}
