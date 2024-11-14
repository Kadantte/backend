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
}
