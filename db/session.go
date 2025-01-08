package db

import (
	"time"

	"morbo/context"
	"morbo/errors"
)

func (db *DB) cleanupStaleSessions(ctx context.Context) error {
	db.log.Info.Println("cleaning up stale sessions")

	query := `DELETE FROM sessions WHERE last_access < NOW() - INTERVAL '30 days';`
	result, err := db.Pool.Exec(ctx, query)
	if err != nil {
		db.log.Error.Println(err)
		db.log.Error.Println("failed to run the query to clean up stale sessions")
		return errors.Err
	}

	rowsAffected := result.RowsAffected()
	db.log.Info.Println("deleted", rowsAffected, "stale sessions")

	return nil
}

func (db *DB) StartPeriodicStaleSessionsCleanup(ctx context.Context) {
	wg := context.GetWaitGroup(ctx)

	db.log.Info.Println("starting periodic stale sessions cleanup")

	wg.Add(1)
	go func() {
		defer wg.Done()

		ticker := time.NewTicker(time.Hour)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				if err := db.cleanupStaleSessions(ctx); err != nil {
					db.log.Error.Println("failed to clean up stale sessions")
				}
			case <-ctx.Done():
				db.log.Info.Println("stopping periodic stale sessions cleanup")
				return
			}
		}
	}()
}
