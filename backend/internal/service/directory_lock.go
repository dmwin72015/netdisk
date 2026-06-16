package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/netdisk/server/internal/model"
)

type LockedDirectory struct {
	ID               int64  `json:"id"`
	Slug             string `json:"slug"`
	Name             string `json:"name"`
	LockPasswordHash string `json:"-"`
}

func ensureFileUnlocked(ctx context.Context, pg *pgxpool.Pool, userID int64, sessionID string, fileID int64) error {
	locked, err := findLockedAncestor(ctx, pg, userID, sessionID, fileID)
	if err != nil {
		return err
	}
	if locked != nil {
		return model.ErrDirectoryLocked
	}
	return nil
}

// filterLockedFileIDs takes a list of file IDs and returns only those that are NOT
// inside a locked directory for which the session has no active unlock.
// This replaces N+1 calls to findLockedAncestor with a single batch query.
func filterLockedFileIDs(ctx context.Context, pg *pgxpool.Pool, userID int64, sessionID string, fileIDs []int64) ([]int64, error) {
	if len(fileIDs) == 0 {
		return nil, nil
	}
	if sessionID == "" {
		sessionID = "__missing_session__"
	}

	// Collect all locked ancestor directories in one query.
	rows, err := pg.Query(ctx, `
		WITH locked_dirs AS (
			SELECT DISTINCT ld.id
			FROM user_files ld
			WHERE ld.user_id = $1
			  AND ld.is_dir = TRUE
			  AND ld.lock_password_hash IS NOT NULL
			  AND NOT EXISTS (
				SELECT 1 FROM user_directory_unlocks u
				WHERE u.user_id = $1
				  AND u.directory_id = ld.id
				  AND u.session_id = $2
				  AND (u.expires_at IS NULL OR u.expires_at > NOW())
			  )
		),
		file_ancestors AS (
			SELECT f.id AS file_id, f.parent_id, 0 AS depth
			FROM user_files f
			WHERE f.id = ANY($3::bigint[]) AND f.user_id = $1
			UNION ALL
			SELECT fa.file_id, p.parent_id, fa.depth + 1
			FROM user_files p
			JOIN file_ancestors fa ON fa.parent_id = p.id
			WHERE p.user_id = $1 AND fa.depth < 50
		)
		SELECT DISTINCT fa.file_id
		FROM file_ancestors fa
		JOIN locked_dirs ld ON ld.id = fa.parent_id
	`, userID, sessionID, fileIDs)
	if err != nil {
		return nil, fmt.Errorf("batch filter locked files: %w", err)
	}
	defer rows.Close()

	lockedSet := make(map[int64]bool)
	for rows.Next() {
		var fid int64
		if err := rows.Scan(&fid); err != nil {
			return nil, fmt.Errorf("scan locked file id: %w", err)
		}
		lockedSet[fid] = true
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	visible := fileIDs[:0]
	for _, fid := range fileIDs {
		if !lockedSet[fid] {
			visible = append(visible, fid)
		}
	}
	return visible, nil
}

func findLockedAncestor(ctx context.Context, pg *pgxpool.Pool, userID int64, sessionID string, fileID int64) (*LockedDirectory, error) {
	if sessionID == "" {
		sessionID = "__missing_session__"
	}
	var locked LockedDirectory
	err := pg.QueryRow(ctx, `
		WITH RECURSIVE ancestors AS (
			SELECT id, slug, file_name, parent_id, lock_password_hash, 0 AS depth
			FROM user_files
			WHERE id = $2 AND user_id = $1
			UNION ALL
			SELECT p.id, p.slug, p.file_name, p.parent_id, p.lock_password_hash, a.depth + 1
			FROM user_files p
			JOIN ancestors a ON a.parent_id = p.id
			WHERE p.user_id = $1 AND a.depth < 50
		)
		SELECT a.id, a.slug, a.file_name, a.lock_password_hash
		FROM ancestors a
		WHERE a.lock_password_hash IS NOT NULL
		  AND NOT EXISTS (
			SELECT 1 FROM user_directory_unlocks u
			WHERE u.user_id = $1
			  AND u.directory_id = a.id
			  AND u.session_id = $3
			  AND (u.expires_at IS NULL OR u.expires_at > NOW())
		  )
		ORDER BY a.depth ASC
		LIMIT 1
	`, userID, fileID, sessionID).Scan(&locked.ID, &locked.Slug, &locked.Name, &locked.LockPasswordHash)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("find locked ancestor: %w", err)
	}
	return &locked, nil
}
