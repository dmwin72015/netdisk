package service

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/zerolog"

	"github.com/netdisk/server/internal/config"
	"github.com/netdisk/server/internal/db/sqlc"
	"github.com/netdisk/server/internal/storage"
)

// TrashWorker periodically purges files that have been in the trash for more than retention_days.
type TrashWorker struct {
	queries *sqlc.Queries
	pg      *pgxpool.Pool
	store   *storage.Local
	logger  zerolog.Logger
	cfg     *config.Config
}

// NewTrashWorker creates a new TrashWorker.
func NewTrashWorker(queries *sqlc.Queries, pg *pgxpool.Pool, store *storage.Local, logger zerolog.Logger, cfg *config.Config) *TrashWorker {
	return &TrashWorker{
		queries: queries,
		pg:      pg,
		store:   store,
		logger:  logger,
		cfg:     cfg,
	}
}

// Start begins the periodic purge loop. It runs every poll_interval until ctx is cancelled.
func (w *TrashWorker) Start(ctx context.Context) {
	w.logger.Info().Msg("trash worker started")

	ticker := time.NewTicker(w.cfg.Trash.PollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			w.logger.Info().Msg("trash worker stopped")
			return
		case <-ticker.C:
			w.purgeExpired(ctx)
		}
	}
}

func (w *TrashWorker) purgeExpired(ctx context.Context) {
	files, err := w.queries.GetExpiredTrashedFiles(ctx)
	if err != nil {
		w.logger.Error().Err(err).Msg("get expired trashed files")
		return
	}

	if len(files) == 0 {
		return
	}

	// Group by user_id, summing file_size for non-directory files.
	type userStats struct {
		reclaimedSize int64
	}
	userMap := make(map[int64]*userStats)
	var ids []int64
	var physicalIDs []int64

	for _, f := range files {
		ids = append(ids, f.ID)

		if !f.IsDir && f.PhysicalFileID.Valid {
			us, ok := userMap[f.UserID]
			if !ok {
				us = &userStats{}
				userMap[f.UserID] = us
			}
			us.reclaimedSize += f.FileSize
			physicalIDs = append(physicalIDs, f.PhysicalFileID.Int64)
		}
	}

	// Batch delete all expired rows.
	if _, err := w.pg.Exec(ctx, "DELETE FROM user_files WHERE id = ANY($1)", ids); err != nil {
		w.logger.Error().Err(err).Msg("batch delete expired trashed files")
		return
	}

	// Reclaim storage for each user.
	for userID, us := range userMap {
		if us.reclaimedSize > 0 {
			if _, err := w.queries.AtomicIncrementStorage(ctx, sqlc.AtomicIncrementStorageParams{
				UserID:      userID,
				StorageUsed: -us.reclaimedSize,
			}); err != nil {
				w.logger.Error().Err(err).Int64("user_id", userID).Msg("reclaim storage")
			}
		}
	}

	// Clean up unreferenced physical files.
	for _, pfID := range physicalIDs {
		refCount, err := w.queries.CountReferencesByFileID(ctx, pgtype.Int8{Int64: pfID, Valid: true})
		if err != nil {
			w.logger.Error().Err(err).Int64("physical_file_id", pfID).Msg("count references")
			continue
		}
		if refCount == 0 {
			pf, err := w.queries.GetPhysicalFileByID(ctx, pfID)
			if err != nil {
				w.logger.Error().Err(err).Int64("physical_file_id", pfID).Msg("get physical file")
				continue
			}
			if err := w.store.Delete(pf.FileHash); err != nil {
				w.logger.Error().Err(err).Str("file_hash", pf.FileHash).Msg("delete physical file from disk")
			}
			if err := w.queries.DeletePhysicalFile(ctx, pf.ID); err != nil {
				w.logger.Error().Err(err).Int64("physical_file_id", pf.ID).Msg("delete physical file from db")
			}
		}
	}

	w.logger.Info().Int("count", len(files)).Msg(fmt.Sprintf("purged %d expired trashed files", len(files)))
}
