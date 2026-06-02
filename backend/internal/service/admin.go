package service

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/zerolog"

	"github.com/netdisk/server/internal/db/sqlc"
)

type AdminService struct {
	queries *sqlc.Queries
	pg      *pgxpool.Pool
	logger  zerolog.Logger
}

func NewAdminService(queries *sqlc.Queries, pg *pgxpool.Pool, logger zerolog.Logger) *AdminService {
	return &AdminService{queries: queries, pg: pg, logger: logger}
}

func (s *AdminService) Queries() *sqlc.Queries { return s.queries }

type AdminUserItem struct {
	ID              string             `json:"id"`
	Slug            string             `json:"slug"`
	Username        string             `json:"username"`
	Email           string             `json:"email"`
	Role            string             `json:"role"`
	RegisterMethod  string             `json:"registerMethod"`
	Status          int16              `json:"status"`
	UsedBytes       int64              `json:"usedBytes"`
	BaseBytes       int64              `json:"baseBytes"`
	MemberBonusBytes int64             `json:"memberBonusBytes"`
	PackBytes       int64              `json:"packBytes"`
	TotalBytes      int64              `json:"totalBytes"`
	CreatedAt       int64              `json:"createdAt"`
	Profile         *AdminUserProfile  `json:"profile,omitempty"`
	OAuthAccounts   []AdminOAuthAccount `json:"oauthAccounts,omitempty"`
}

type AdminUserProfile struct {
	DisplayName string `json:"displayName"`
	AvatarURL   string `json:"avatarUrl"`
	Bio         string `json:"bio"`
}

type AdminOAuthAccount struct {
	Provider          string `json:"provider"`
	ProviderAccountID string `json:"providerAccountId"`
	CreatedAt         int64  `json:"createdAt"`
}

type AdminFileItem struct {
	ID           string `json:"id"`
	Slug         string `json:"slug"`
	UserID       string `json:"userId"`
	Username     string `json:"username"`
	FileName     string `json:"fileName"`
	IsDir        bool   `json:"isDir"`
	FileSize     int64  `json:"fileSize"`
	MimeType     string `json:"mimeType"`
	FileCategory string `json:"fileCategory"`
	IsTrashed    bool   `json:"isTrashed"`
	IsStarred    bool   `json:"isStarred"`
	CreatedAt    int64  `json:"createdAt"`
	UpdatedAt    int64  `json:"updatedAt"`
}

func (s *AdminService) ListUsers(ctx context.Context, limit, offset int) ([]AdminUserItem, int, error) {
	var total int
	err := s.pg.QueryRow(ctx, "SELECT COUNT(*) FROM users").Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("count users: %w", err)
	}

	rows, err := s.pg.Query(ctx, `
		SELECT u.id, u.slug, u.username, u.email, u.role, u.register_method, u.status,
		       COALESCE(ss.storage_used, 0), COALESCE(ss.storage_quota, 0),
		       EXTRACT(EPOCH FROM u.created_at)::bigint
		FROM users u
		LEFT JOIN user_storage_stats ss ON ss.user_id = u.id
		ORDER BY u.created_at DESC
		LIMIT $1 OFFSET $2
	`, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("list users: %w", err)
	}
	defer rows.Close()

	var items []AdminUserItem
	for rows.Next() {
		var item AdminUserItem
		var id int64
		if err := rows.Scan(
			&id, &item.Slug, &item.Username, &item.Email,
			&item.Role, &item.RegisterMethod, &item.Status,
			&item.UsedBytes, &item.TotalBytes,
			&item.CreatedAt,
		); err != nil {
			return nil, 0, fmt.Errorf("scan user: %w", err)
		}
		item.BaseBytes = item.TotalBytes
		item.ID = fmt.Sprintf("%d", id)
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, err
	}

	return items, total, nil
}

func (s *AdminService) GetUser(ctx context.Context, id int64) (*AdminUserItem, error) {
	var item AdminUserItem
	var rawID int64
	err := s.pg.QueryRow(ctx, `
		SELECT u.id, u.slug, u.username, u.email, u.role, u.register_method, u.status,
		       COALESCE(ss.storage_used, 0), COALESCE(ss.storage_quota, 0),
		       EXTRACT(EPOCH FROM u.created_at)::bigint
		FROM users u
		LEFT JOIN user_storage_stats ss ON ss.user_id = u.id
		WHERE u.id = $1
	`, id).Scan(
		&rawID, &item.Slug, &item.Username, &item.Email,
		&item.Role, &item.RegisterMethod, &item.Status,
		&item.UsedBytes, &item.TotalBytes,
		&item.CreatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("get user by id: %w", err)
	}
	item.ID = fmt.Sprintf("%d", rawID)
	item.BaseBytes = item.TotalBytes

	var profile AdminUserProfile
	err = s.pg.QueryRow(ctx, `
		SELECT COALESCE(display_name, ''), COALESCE(avatar_path, ''), COALESCE(bio, '')
		FROM user_profiles WHERE user_id = $1
	`, id).Scan(&profile.DisplayName, &profile.AvatarURL, &profile.Bio)
	if err == nil {
		item.Profile = &profile
	}

	oauthRows, err := s.pg.Query(ctx, `
		SELECT provider, provider_account_id, EXTRACT(EPOCH FROM created_at)::bigint
		FROM user_oauth_accounts WHERE user_id = $1
	`, id)
	if err != nil {
		return &item, nil
	}
	defer oauthRows.Close()
	for oauthRows.Next() {
		var oa AdminOAuthAccount
		if err := oauthRows.Scan(&oa.Provider, &oa.ProviderAccountID, &oa.CreatedAt); err != nil {
			continue
		}
		item.OAuthAccounts = append(item.OAuthAccounts, oa)
	}

	return &item, nil
}

func (s *AdminService) UpdateRole(ctx context.Context, id int64, role string) (*AdminUserItem, error) {
	_, err := s.pg.Exec(ctx, "UPDATE users SET role = $2, updated_at = NOW() WHERE id = $1", id, role)
	if err != nil {
		return nil, fmt.Errorf("update role: %w", err)
	}
	return s.GetUser(ctx, id)
}

func (s *AdminService) UpdateStorageBase(ctx context.Context, id int64, baseBytes int64) (*AdminUserItem, error) {
	_, err := s.pg.Exec(ctx, `
		UPDATE user_storage_stats
		SET storage_quota = $2, updated_at = NOW()
		WHERE user_id = $1
	`, id, baseBytes)
	if err != nil {
		return nil, fmt.Errorf("update storage base: %w", err)
	}
	return s.GetUser(ctx, id)
}

func (s *AdminService) DeleteUser(ctx context.Context, id int64) error {
	tx, err := s.pg.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback(ctx)

	tag, err := tx.Exec(ctx, "DELETE FROM user_oauth_accounts WHERE user_id = $1", id)
	if err != nil {
		return fmt.Errorf("delete oauth accounts: %w", err)
	}
	_ = tag

	_, err = tx.Exec(ctx, "DELETE FROM user_storage_stats WHERE user_id = $1", id)
	if err != nil {
		return fmt.Errorf("delete storage stats: %w", err)
	}

	_, err = tx.Exec(ctx, "DELETE FROM user_profiles WHERE user_id = $1", id)
	if err != nil {
		return fmt.Errorf("delete profile: %w", err)
	}

	// Delete associated user files (soft-delete by updating is_trashed or hard delete)
	_, err = tx.Exec(ctx, "DELETE FROM user_files WHERE user_id = $1", id)
	if err != nil {
		return fmt.Errorf("delete files: %w", err)
	}

	_, err = tx.Exec(ctx, "DELETE FROM refresh_tokens WHERE user_id = $1", id)
	if err != nil {
		return fmt.Errorf("delete refresh tokens: %w", err)
	}

	tag, err = tx.Exec(ctx, "DELETE FROM users WHERE id = $1", id)
	if err != nil {
		return fmt.Errorf("delete user: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("user not found")
	}

	return tx.Commit(ctx)
}

func (s *AdminService) ListFiles(ctx context.Context, limit, offset int) ([]AdminFileItem, int, error) {
	var total int
	err := s.pg.QueryRow(ctx, "SELECT COUNT(*) FROM user_files").Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("count files: %w", err)
	}

	rows, err := s.pg.Query(ctx, `
		SELECT uf.id, uf.slug, uf.user_id, u.username, uf.file_name, uf.is_dir,
		       uf.file_size, COALESCE(uf.mime_type, ''), COALESCE(uf.file_category, ''),
		       uf.is_trashed, uf.is_starred,
		       EXTRACT(EPOCH FROM uf.created_at)::bigint,
		       EXTRACT(EPOCH FROM uf.updated_at)::bigint
		FROM user_files uf
		JOIN users u ON u.id = uf.user_id
		ORDER BY uf.created_at DESC
		LIMIT $1 OFFSET $2
	`, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("list files: %w", err)
	}
	defer rows.Close()

	var items []AdminFileItem
	for rows.Next() {
		var item AdminFileItem
		var fileID, userID int64
		if err := rows.Scan(
			&fileID, &item.Slug, &userID, &item.Username,
			&item.FileName, &item.IsDir, &item.FileSize,
			&item.MimeType, &item.FileCategory,
			&item.IsTrashed, &item.IsStarred,
			&item.CreatedAt, &item.UpdatedAt,
		); err != nil {
			return nil, 0, fmt.Errorf("scan file: %w", err)
		}
		item.ID = fmt.Sprintf("%d", fileID)
		item.UserID = fmt.Sprintf("%d", userID)
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, err
	}

	return items, total, nil
}
