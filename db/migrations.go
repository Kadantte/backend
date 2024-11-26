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
	{
		version: 3,
		sql: `
			INSERT INTO users
				(id, username, password)
			VALUES
				(1, 'admin', '$2a$10$X0W3DOiy9dUP0F9xOX5o.uxckTsdpnzMJLiMYqE2kHnRIDYfWDfqC');

			UPDATE schema_version SET version = 3;
		`,
	},
}
