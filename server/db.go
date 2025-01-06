package server

import (
	"morbo/errors"
	"net/http"

	"github.com/jackc/pgx/v5"
)

func (conn *Connection) QueryRow(query string, args ...any) pgx.Row {
	return conn.db.Pool.QueryRow(conn.ctx, query, args...)
}

func (conn *Connection) ScanRow(row pgx.Row, dest ...any) error {
	if !conn.ContextAlive() {
		return errors.Error
	}
	err := row.Scan(dest...)
	if err != nil {
		switch err {
		case pgx.ErrNoRows:
		default:
			conn.log.Error.Println(err)
			conn.DistinctError(
				"failed to query for a row",
				"internal server error",
				http.StatusInternalServerError,
			)
		}
	}
	return err
}

func (conn *Connection) Exec(query string, args ...any) error {
	if !conn.ContextAlive() {
		return errors.Error
	}

	if _, err := conn.db.Pool.Exec(conn.ctx, query, args...); err != nil {
		conn.log.Error.Println(err)
		conn.DistinctError(
			"failed to execute the statement",
			"internal server error",
			http.StatusInternalServerError,
		)
		return errors.Error
	}
	return nil
}
