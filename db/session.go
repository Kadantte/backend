package db

import (
	"time"

	"morbo/context"
	"morbo/errors"
	"morbo/log"
)

func (db *DB) cleanupStaleSessions(ctx context.Context) error {
	log.Info.Println("cleaning up stale sessions")

	query := `DELETE FROM sessions WHERE last_access < NOW() - INTERVAL '30 days';`
	result, err := db.Pool.Exec(ctx, query)
	if err != nil {
		log.Error.Println(err)
		log.Error.Println("failed to run the query to clean up stale sessions")
		return errors.Error
	}

	rowsAffected := result.RowsAffected()
	log.Info.Println("deleted", rowsAffected, "stale sessions")

	return nil
}

func (db *DB) StartPeriodicStaleSessionsCleanup(ctx context.Context) {
	wg := context.GetWaitGroup(ctx)

	log.Info.Println("starting periodic stale sessions cleanup")

	wg.Add(1)
	go func() {
		defer wg.Done()

		ticker := time.NewTicker(time.Hour)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				if err := db.cleanupStaleSessions(ctx); err != nil {
					log.Error.Println("failed to clean up stale sessions")
				}
			case <-ctx.Done():
				log.Info.Println("stopping periodic stale sessions cleanup")
				return
			}
		}
	}()
}
