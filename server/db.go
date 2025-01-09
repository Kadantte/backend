package server

import (
	"net/http"

	"morbo/context"
	"morbo/errors"

	"github.com/jackc/pgx/v5"
)

func (conn *Connection) QueryRow(ctx context.Context, query string, args ...any) pgx.Row {
	return conn.db.Pool.QueryRow(ctx, query, args...)
}

func (conn *Connection) ScanRow(ctx context.Context, row pgx.Row, dest ...any) error {
	if !conn.ContextAlive(ctx) {
		return errors.Err
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

func (conn *Connection) Exec(ctx context.Context, query string, args ...any) error {
	if !conn.ContextAlive(ctx) {
		return errors.Err
	}

	if _, err := conn.db.Pool.Exec(ctx, query, args...); err != nil {
		conn.log.Error.Println(err)
		conn.DistinctError(
			"failed to execute the statement",
			"internal server error",
			http.StatusInternalServerError,
		)
		return errors.Err
	}

	return nil
}
