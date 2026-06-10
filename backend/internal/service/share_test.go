package service

import (
	"context"
	"errors"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/netdisk/server/internal/model"
)

func TestGetPublicInfoReturnsNotFoundForDeletedShare(t *testing.T) {
	dsn := os.Getenv("NETDISK_TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("NETDISK_TEST_DATABASE_URL is required for share integration tests")
	}

	ctx := context.Background()
	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		t.Fatalf("connect test database: %v", err)
	}
	defer pool.Close()

	shareSlug := fmt.Sprintf("delshare%013d", time.Now().UnixNano()%10000000000000)
	userSlug := fmt.Sprintf("user%017d", time.Now().UnixNano()%100000000000000000)
	fileSlug := fmt.Sprintf("file%017d", time.Now().UnixNano()%100000000000000000)
	physicalSlug := fmt.Sprintf("phys%017d", time.Now().UnixNano()%100000000000000000)
	fileHash := fmt.Sprintf("%064x", time.Now().UnixNano())

	var userID int64
	if err := pool.QueryRow(ctx, `
		INSERT INTO users (slug, username, email, password_hash)
		VALUES ($1, $2, $3, 'hash')
		RETURNING id
	`, userSlug, userSlug, userSlug+"@example.test").Scan(&userID); err != nil {
		t.Fatalf("insert user: %v", err)
	}
	defer pool.Exec(ctx, `DELETE FROM users WHERE id = $1`, userID)

	var physicalFileID int64
	if err := pool.QueryRow(ctx, `
		INSERT INTO physical_files (slug, file_hash, pre_hash, file_size, mime_type, storage_path)
		VALUES ($1, $2, $3, 12, 'text/plain', $4)
		RETURNING id
	`, physicalSlug, fileHash, fileHash, fileHash).Scan(&physicalFileID); err != nil {
		t.Fatalf("insert physical file: %v", err)
	}
	defer pool.Exec(ctx, `DELETE FROM physical_files WHERE id = $1`, physicalFileID)

	var userFileID int64
	if err := pool.QueryRow(ctx, `
		INSERT INTO user_files (slug, user_id, physical_file_id, file_name, is_dir, file_size, mime_type)
		VALUES ($1, $2, $3, 'deleted-share.txt', FALSE, 12, 'text/plain')
		RETURNING id
	`, fileSlug, userID, physicalFileID).Scan(&userFileID); err != nil {
		t.Fatalf("insert user file: %v", err)
	}
	defer pool.Exec(ctx, `DELETE FROM user_files WHERE id = $1`, userFileID)

	var shareID int64
	if err := pool.QueryRow(ctx, `
		INSERT INTO file_shares (slug, user_id, deleted_at)
		VALUES ($1, $2, NOW())
		RETURNING id
	`, shareSlug, userID).Scan(&shareID); err != nil {
		t.Fatalf("insert deleted share: %v", err)
	}
	defer pool.Exec(ctx, `DELETE FROM file_shares WHERE id = $1`, shareID)

	var shareItemID int64
	if err := pool.QueryRow(ctx, `
		INSERT INTO file_share_items (share_id, file_id)
		VALUES ($1, $2)
		RETURNING id
	`, shareID, userFileID).Scan(&shareItemID); err != nil {
		t.Fatalf("insert share item: %v", err)
	}
	defer pool.Exec(ctx, `DELETE FROM file_share_items WHERE id = $1`, shareItemID)

	service := NewShareService(pool, nil)
	_, err = service.GetPublicInfo(ctx, shareSlug)
	if !errors.Is(err, model.ErrNotFound) {
		t.Fatalf("GetPublicInfo() error = %v, want %v", err, model.ErrNotFound)
	}
}
