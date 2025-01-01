package server

import (
	"morbo/errors"

	"github.com/jackc/pgx/v5"
)

func (conn *Connection) QueryRow(query string, args ...any) pgx.Row {
	return conn.db.Pool.QueryRow(conn.ctx, query, args...)
}

func (conn *Connection) ScanRow(row pgx.Row, dest ...any) error {
	if !conn.ContextAlive() {
		return errors.Error
	}
	return row.Scan(dest...)
}

func (conn *Connection) Exec(query string, args ...any) error {
	if !conn.ContextAlive() {
		return errors.Error
	}
	_, err := conn.db.Pool.Exec(conn.ctx, query, args...)
	return err
}
