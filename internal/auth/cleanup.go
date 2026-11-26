package auth

import (
	"context"
	"log"
	"time"
)

func RunRefreshSessionCleanup(ctx context.Context, repository *Repository, logger *log.Logger) {
	cleanup := func() {
		cleanupCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
		defer cancel()

		deletedCount, err := repository.DeleteExpiredSessions(cleanupCtx)
		if err != nil {
			logger.Printf("cleanup refresh sessions: %v", err)
			return
		}

		logger.Printf("deleted refresh sessions: %d", deletedCount)
	}

	cleanup()

	ticker := time.NewTicker(24 * time.Hour)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			cleanup()
		}

	}
}
