package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"
	"syscall"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/zerolog"
	"golang.org/x/crypto/bcrypt"

	"github.com/netdisk/server/internal/db/sqlc"
	"github.com/netdisk/server/internal/i18n"
	"github.com/netdisk/server/internal/model"
	"github.com/netdisk/server/internal/storage"
)

type AdminService struct {
	queries     *sqlc.Queries
	pg          *pgxpool.Pool
	logger      zerolog.Logger
	storageRoot string
	filesDir    string
	configSvc   *SystemConfigService
}

func NewAdminService(queries *sqlc.Queries, pg *pgxpool.Pool, logger zerolog.Logger, storageRoot string, filesDir string, configSvc *SystemConfigService) *AdminService {
	return &AdminService{queries: queries, pg: pg, logger: logger, storageRoot: storageRoot, filesDir: filesDir, configSvc: configSvc}
}

func (s *AdminService) Queries() *sqlc.Queries { return s.queries }

type AdminUserItem struct {
	ID               string              `json:"id"`
	Slug             string              `json:"slug"`
	Username         string              `json:"username"`
	Email            string              `json:"email"`
	Role             string              `json:"role"`
	RegisterMethod   string              `json:"registerMethod"`
	Status           int16               `json:"status"`
	UsedBytes        int64               `json:"usedBytes"`
	BaseBytes        int64               `json:"baseBytes"`
	MemberBonusBytes int64               `json:"memberBonusBytes"`
	PackBytes        int64               `json:"packBytes"`
	TotalBytes       int64               `json:"totalBytes"`
	CreatedAt        int64               `json:"createdAt"`
	Profile          *AdminUserProfile   `json:"profile,omitempty"`
	OAuthAccounts    []AdminOAuthAccount `json:"oauthAccounts,omitempty"`
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
	FileHash     string `json:"fileHash"`
	IsTrashed    bool   `json:"isTrashed"`
	IsStarred    bool   `json:"isStarred"`
	CreatedAt    int64  `json:"createdAt"`
	UpdatedAt    int64  `json:"updatedAt"`
}

type AdminDashboardStats struct {
	TotalUsers     int64 `json:"totalUsers"`
	TotalFiles     int64 `json:"totalFiles"`
	TotalStorage   int64 `json:"totalStorage"`
	StorageUsed    int64 `json:"storageUsed"`
	NewTodayUsers  int64 `json:"newTodayUsers"`
	NewTodayFiles  int64 `json:"newTodayFiles"`
	DiskTotal      int64 `json:"diskTotal"`
	DiskUsed       int64 `json:"diskUsed"`
	DiskFree       int64 `json:"diskFree"`
}

func (s *AdminService) DashboardStats(ctx context.Context) (*AdminDashboardStats, error) {
	var stats AdminDashboardStats

	err := s.pg.QueryRow(ctx, "SELECT COUNT(*) FROM users").Scan(&stats.TotalUsers)
	if err != nil {
		return nil, fmt.Errorf("count users: %w", err)
	}

	err = s.pg.QueryRow(ctx, "SELECT COUNT(*) FROM user_files WHERE is_trashed = false").Scan(&stats.TotalFiles)
	if err != nil {
		return nil, fmt.Errorf("count files: %w", err)
	}

	err = s.pg.QueryRow(ctx, "SELECT COALESCE(SUM(storage_quota), 0) FROM user_storage_stats").Scan(&stats.TotalStorage)
	if err != nil {
		return nil, fmt.Errorf("sum storage quota: %w", err)
	}

	err = s.pg.QueryRow(ctx, "SELECT COALESCE(SUM(storage_used), 0) FROM user_storage_stats").Scan(&stats.StorageUsed)
	if err != nil {
		return nil, fmt.Errorf("sum storage used: %w", err)
	}

	err = s.pg.QueryRow(ctx, "SELECT COUNT(*) FROM users WHERE created_at >= CURRENT_DATE").Scan(&stats.NewTodayUsers)
	if err != nil {
		return nil, fmt.Errorf("count new users today: %w", err)
	}

	err = s.pg.QueryRow(ctx, "SELECT COUNT(*) FROM user_files WHERE created_at >= CURRENT_DATE AND is_trashed = false").Scan(&stats.NewTodayFiles)
	if err != nil {
		return nil, fmt.Errorf("count new files today: %w", err)
	}

	if s.storageRoot != "" {
		var stat syscall.Statfs_t
		if err := syscall.Statfs(s.storageRoot, &stat); err == nil {
			bsize := int64(stat.Bsize)
			stats.DiskTotal = int64(stat.Blocks) * bsize
			stats.DiskFree = int64(stat.Bavail) * bsize
			stats.DiskUsed = stats.DiskTotal - int64(stat.Bfree)*bsize
		}
	}

	return &stats, nil
}

type AdminListUsersParams struct {
	Limit          int
	Offset         int
	Search         string
	Role           string
	Sort           string
	Slug           string
	Email          string
	RegisterMethod string
	Status         int16
	CreatedAfter   *time.Time
	CreatedBefore  *time.Time
}

func (s *AdminService) ListUsers(ctx context.Context, params AdminListUsersParams) ([]AdminUserItem, int, error) {
	where := "TRUE"
	args := []any{}
	argIdx := 1

	if params.Search != "" {
		where += fmt.Sprintf(" AND (u.username ILIKE '%%' || $%d || '%%' OR u.email ILIKE '%%' || $%d || '%%')", argIdx, argIdx)
		args = append(args, params.Search)
		argIdx++
	}
	if params.Role != "" {
		where += fmt.Sprintf(" AND u.role = $%d", argIdx)
		args = append(args, params.Role)
		argIdx++
	}
	if params.CreatedAfter != nil {
		where += fmt.Sprintf(" AND u.created_at >= $%d", argIdx)
		args = append(args, params.CreatedAfter)
		argIdx++
	}
	if params.CreatedBefore != nil {
		where += fmt.Sprintf(" AND u.created_at <= $%d", argIdx)
		args = append(args, params.CreatedBefore)
		argIdx++
	}
	if params.Slug != "" {
		where += fmt.Sprintf(" AND u.slug ILIKE '%%' || $%d || '%%'", argIdx)
		args = append(args, params.Slug)
		argIdx++
	}
	if params.Email != "" {
		where += fmt.Sprintf(" AND u.email ILIKE '%%' || $%d || '%%'", argIdx)
		args = append(args, params.Email)
		argIdx++
	}
	if params.RegisterMethod != "" {
		where += fmt.Sprintf(" AND u.register_method = $%d", argIdx)
		args = append(args, params.RegisterMethod)
		argIdx++
	}
	if params.Status != 0 {
		where += fmt.Sprintf(" AND u.status = $%d", argIdx)
		args = append(args, params.Status)
		argIdx++
	}

	var total int
	countQ := "SELECT COUNT(*) FROM users u WHERE " + where
	err := s.pg.QueryRow(ctx, countQ, args...).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("count users: %w", err)
	}

	orderBy := "u.created_at DESC"
	switch params.Sort {
	case "username":
		orderBy = "u.username ASC"
	case "-username":
		orderBy = "u.username DESC"
	case "storage":
		orderBy = "COALESCE(ss.storage_used, 0) DESC"
	case "-storage":
		orderBy = "COALESCE(ss.storage_used, 0) ASC"
	case "created_at":
		orderBy = "u.created_at ASC"
	case "-created_at":
		orderBy = "u.created_at DESC"
	default:
		orderBy = "u.created_at DESC"
	}

	q := fmt.Sprintf(`
		SELECT u.id, u.slug, u.username, u.email, u.role, u.register_method, u.status,
		       COALESCE(ss.storage_used, 0), COALESCE(ss.storage_quota, 0),
		       EXTRACT(EPOCH FROM u.created_at)::bigint
		FROM users u
		LEFT JOIN user_storage_stats ss ON ss.user_id = u.id
		WHERE %s
		ORDER BY %s
		LIMIT $%d OFFSET $%d
	`, where, orderBy, argIdx, argIdx+1)

	args = append(args, params.Limit, params.Offset)
	rows, err := s.pg.Query(ctx, q, args...)
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

func (s *AdminService) CreateUser(ctx context.Context, username, email, password, role string) (*AdminUserItem, error) {
	tx, err := s.pg.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback(ctx)

	hash, err := bcrypt.GenerateFromPassword([]byte(password), 12)
	if err != nil {
		return nil, fmt.Errorf("hash password: %w", err)
	}

	var userID int64
	err = tx.QueryRow(ctx, `
		INSERT INTO users (username, email, password_hash, role, register_method, status)
		VALUES ($1, $2, $3, $4, 'admin', 1)
		RETURNING id
	`, username, email, hash, role).Scan(&userID)
	if err != nil {
		if strings.Contains(err.Error(), "unique") || strings.Contains(err.Error(), "duplicate") {
			return nil, fmt.Errorf("username or email already exists")
		}
		return nil, fmt.Errorf("insert user: %w", err)
	}

		quota := int64(500 << 30)
		if v, ok := s.configSvc.Get("default_quota"); ok {
			switch n := v.(type) {
			case int64:
				quota = n
			case float64:
				quota = int64(n)
			}
		}
		_, err = tx.Exec(ctx, `
			INSERT INTO user_storage_stats (user_id, storage_used, storage_quota)
			VALUES ($1, 0, $2)
		`, userID, quota)
	if err != nil {
		return nil, fmt.Errorf("insert storage stats: %w", err)
	}

	_, err = tx.Exec(ctx, `
		INSERT INTO user_profiles (user_id, display_name, avatar_path, bio)
		VALUES ($1, '', '', '')
	`, userID)
	if err != nil {
		return nil, fmt.Errorf("insert profile: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("commit tx: %w", err)
	}

	return s.GetUser(ctx, userID)
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

type AdminListFilesParams struct {
	Limit    int
	Offset   int
	Search   string
	Category string
	Trashed  string
	Sort     string
}

func (s *AdminService) ListFiles(ctx context.Context, params AdminListFilesParams) ([]AdminFileItem, int, error) {
	where := "TRUE"
	args := []any{}
	argIdx := 1

	if params.Search != "" {
		where += fmt.Sprintf(" AND (uf.file_name ILIKE '%%' || $%d || '%%')", argIdx)
		args = append(args, params.Search)
		argIdx++
	}
	if params.Category != "" {
		where += fmt.Sprintf(" AND uf.file_category = $%d", argIdx)
		args = append(args, params.Category)
		argIdx++
	}
	switch params.Trashed {
	case "only":
		where += " AND uf.is_trashed = true"
	case "no":
		where += " AND uf.is_trashed = false"
	}

	var total int
	countQ := "SELECT COUNT(*) FROM user_files uf WHERE " + where
	err := s.pg.QueryRow(ctx, countQ, args...).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("count files: %w", err)
	}

	orderBy := "uf.created_at DESC"
	switch params.Sort {
	case "name":
		orderBy = "uf.file_name ASC"
	case "-name":
		orderBy = "uf.file_name DESC"
	case "size":
		orderBy = "uf.file_size DESC"
	case "-size":
		orderBy = "uf.file_size ASC"
	case "created_at":
		orderBy = "uf.created_at ASC"
	case "-created_at":
		orderBy = "uf.created_at DESC"
	default:
		orderBy = "uf.created_at DESC"
	}

	q := fmt.Sprintf(`
		SELECT uf.id, uf.slug, uf.user_id, u.username, uf.file_name, uf.is_dir,
		       uf.file_size, COALESCE(uf.mime_type, ''), COALESCE(uf.file_category, ''),
		       COALESCE(pf.file_hash, ''), uf.is_trashed, uf.is_starred,
		       EXTRACT(EPOCH FROM uf.created_at)::bigint,
		       EXTRACT(EPOCH FROM uf.updated_at)::bigint
		FROM user_files uf
		JOIN users u ON u.id = uf.user_id
		LEFT JOIN physical_files pf ON pf.id = uf.physical_file_id
		WHERE %s
		ORDER BY %s
		LIMIT $%d OFFSET $%d
	`, where, orderBy, argIdx, argIdx+1)

	args = append(args, params.Limit, params.Offset)
	rows, err := s.pg.Query(ctx, q, args...)
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
			&item.MimeType, &item.FileCategory, &item.FileHash,
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

type AdminPhysicalFileItem struct {
	ID             int64  `json:"id"`
	Slug           string `json:"slug"`
	HashAlgo       string `json:"hashAlgo"`
	FileHash       string `json:"fileHash"`
	PreHash        string `json:"preHash"`
	FileSize       int64  `json:"fileSize"`
	MimeType       string `json:"mimeType"`
	StoragePath    string `json:"storagePath"`
	Status         string `json:"status"`
	CreatedAt      int64  `json:"createdAt"`
	UserFileCount  int64  `json:"userFileCount"`
	MediaItemCount int64  `json:"mediaItemCount"`
}

type AdminListPhysicalFilesParams struct {
	Limit       int
	Offset      int
	Search      string
	Status      string
	HashFilter  string
	MimeFilter  string
	MinSize     int64
	MaxSize     int64
	CreatedFrom string
	CreatedTo   string
}

func (s *AdminService) ListPhysicalFiles(ctx context.Context, p AdminListPhysicalFilesParams) ([]AdminPhysicalFileItem, int, error) {
	makeText := func(s string) pgtype.Text {
		if s == "" {
			return pgtype.Text{}
		}
		return pgtype.Text{String: s, Valid: true}
	}
	makeInt8 := func(v int64) pgtype.Int8 {
		if v == 0 {
			return pgtype.Int8{}
		}
		return pgtype.Int8{Int64: v, Valid: true}
	}
	makeTime := func(s string) pgtype.Timestamptz {
		if s == "" {
			return pgtype.Timestamptz{}
		}
		t, err := time.Parse(time.RFC3339, s)
		if err == nil {
			return pgtype.Timestamptz{Time: t, Valid: true}
		}
		return pgtype.Timestamptz{}
	}

	params := sqlc.CountPhysicalFilesParams{
		Status:      makeText(p.Status),
		Search:      makeText(p.Search),
		HashFilter:  makeText(p.HashFilter),
		MimeFilter:  makeText(p.MimeFilter),
		MinSize:     makeInt8(p.MinSize),
		MaxSize:     makeInt8(p.MaxSize),
		CreatedFrom: makeTime(p.CreatedFrom),
		CreatedTo:   makeTime(p.CreatedTo),
	}

	total, err := s.queries.CountPhysicalFiles(ctx, params)
	if err != nil {
		return nil, 0, fmt.Errorf("count physical files: %w", err)
	}

	paramsList := sqlc.ListPhysicalFilesParams{
		Status:      makeText(p.Status),
		Search:      makeText(p.Search),
		HashFilter:  makeText(p.HashFilter),
		MimeFilter:  makeText(p.MimeFilter),
		MinSize:     makeInt8(p.MinSize),
		MaxSize:     makeInt8(p.MaxSize),
		CreatedFrom: makeTime(p.CreatedFrom),
		CreatedTo:   makeTime(p.CreatedTo),
		Off:         int32(p.Offset),
		Lim:         int32(p.Limit),
	}

	rows, err := s.queries.ListPhysicalFiles(ctx, paramsList)
	if err != nil {
		return nil, 0, fmt.Errorf("list physical files: %w", err)
	}

	items := make([]AdminPhysicalFileItem, 0, len(rows))
	for _, r := range rows {
		items = append(items, AdminPhysicalFileItem{
			ID:             r.ID,
			Slug:           r.Slug,
			HashAlgo:       r.HashAlgo,
			FileHash:       r.FileHash,
			PreHash:        r.PreHash,
			FileSize:       r.FileSize,
			MimeType:       r.MimeType,
			StoragePath:    r.StoragePath,
			Status:         r.Status,
			CreatedAt:      r.CreatedAt.Time.Unix(),
			UserFileCount:  r.UserFileCount,
			MediaItemCount: r.MediaItemCount,
		})
	}

	return items, int(total), nil
}

type PhysicalFileDetailResult struct {
	PhysicalFile AdminPhysicalFileItem   `json:"physicalFile"`
	UserFiles    []CleanupQueryUserFile `json:"userFiles"`
	TotalUploads int                    `json:"totalUploads"`
	UniqueUsers  int                    `json:"uniqueUsers"`
	FullPath     string                 `json:"fullPath"`
}

func (s *AdminService) PhysicalFileDetail(ctx context.Context, id int64) (*PhysicalFileDetailResult, error) {
	pf, err := s.queries.GetPhysicalFileByID(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, model.ErrNotFound
		}
		return nil, fmt.Errorf("get physical file: %w", err)
	}

	detail := &PhysicalFileDetailResult{
		FullPath: storage.AbsPath(s.storageRoot, pf.FileHash, s.filesDir),
		PhysicalFile: AdminPhysicalFileItem{
			ID:          pf.ID,
			Slug:        pf.Slug,
			HashAlgo:    pf.HashAlgo,
			FileHash:    pf.FileHash,
			PreHash:     pf.PreHash,
			FileSize:    pf.FileSize,
			MimeType:    pf.MimeType,
			StoragePath: pf.StoragePath,
			Status:      pf.Status,
			CreatedAt:   pf.CreatedAt.Time.Unix(),
		},
		UserFiles: []CleanupQueryUserFile{},
	}

	rows, err := s.pg.Query(ctx, `
		SELECT uf.id, uf.slug, uf.user_id, u.username, uf.file_name, uf.file_size,
		       EXTRACT(EPOCH FROM uf.created_at)::bigint
		FROM user_files uf
		JOIN users u ON u.id = uf.user_id
		WHERE uf.physical_file_id = $1
		ORDER BY uf.created_at DESC
	`, pf.ID)
	if err != nil {
		return nil, fmt.Errorf("query user files: %w", err)
	}
	defer rows.Close()

	seen := make(map[int64]bool)
	for rows.Next() {
		var uf CleanupQueryUserFile
		var createdAt int64
		if err := rows.Scan(&uf.ID, &uf.Slug, &uf.UserID, &uf.Username, &uf.FileName, &uf.FileSize, &createdAt); err != nil {
			return nil, fmt.Errorf("scan user file: %w", err)
		}
		uf.CreatedAt = createdAt
		detail.UserFiles = append(detail.UserFiles, uf)
		if !seen[uf.UserID] {
			seen[uf.UserID] = true
			detail.UniqueUsers++
		}
		detail.TotalUploads++
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	// Populate reference counts that the base GetPhysicalFileByID query does not include
	detail.PhysicalFile.UserFileCount = int64(detail.TotalUploads)
	var mediaCount int64
	_ = s.pg.QueryRow(ctx, `SELECT COUNT(*) FROM media_items WHERE physical_file_id = $1`, pf.ID).Scan(&mediaCount)
	detail.PhysicalFile.MediaItemCount = mediaCount

	return detail, nil
}

func (s *AdminService) StorageStats(ctx context.Context) ([]CategoryStat, error) {
	rows, err := s.pg.Query(ctx,
		`SELECT COALESCE(NULLIF(file_category, ''), 'other'), COALESCE(SUM(file_size), 0), COUNT(*)
		 FROM user_files
		 WHERE is_dir = FALSE AND is_trashed = FALSE
		 GROUP BY COALESCE(NULLIF(file_category, ''), 'other')
		 ORDER BY SUM(file_size) DESC`,
	)
	if err != nil {
		return nil, fmt.Errorf("query storage stats: %w", err)
	}
	defer rows.Close()

	statMap := make(map[string]CategoryStat, 8)
	for rows.Next() {
		var cs CategoryStat
		if err := rows.Scan(&cs.Category, &cs.Bytes, &cs.Count); err != nil {
			return nil, fmt.Errorf("scan category stat: %w", err)
		}
		statMap[cs.Category] = cs
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	var trashBytes, trashCount int64
	_ = s.pg.QueryRow(ctx,
		`SELECT COALESCE(SUM(file_size), 0), COUNT(*)
		 FROM user_files
		 WHERE is_dir = FALSE AND is_trashed = TRUE`,
	).Scan(&trashBytes, &trashCount)

	allCategories := []string{"video", "audio", "image", "document", "archive", "other", "trash"}
	stats := make([]CategoryStat, 0, len(allCategories))
	for _, cat := range allCategories {
		var cs CategoryStat
		if cat == "trash" {
			cs = CategoryStat{Category: "trash", Bytes: trashBytes, Count: trashCount}
		} else if c, ok := statMap[cat]; ok {
			cs = c
		} else {
			cs = CategoryStat{Category: cat}
		}
		stats = append(stats, cs)
	}

	return stats, nil
}

func (s *AdminService) DeleteFile(ctx context.Context, fileID int64) error {
	tag, err := s.pg.Exec(ctx, "DELETE FROM user_files WHERE id = $1", fileID)
	if err != nil {
		return fmt.Errorf("delete file: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("file not found")
	}
	return nil
}

func (s *AdminService) RestoreFile(ctx context.Context, fileID int64) error {
	tag, err := s.pg.Exec(ctx, "UPDATE user_files SET is_trashed = false, updated_at = NOW() WHERE id = $1 AND is_trashed = true", fileID)
	if err != nil {
		return fmt.Errorf("restore file: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("file not found or not trashed")
	}
	return nil
}


type CleanupQueryPhysicalFile struct {
	ID          int64  `json:"id"`
	FileHash    string `json:"fileHash"`
	FileSize    int64  `json:"fileSize"`
	StoragePath string `json:"storagePath"`
	MimeType    string `json:"mimeType"`
	FileExists  bool   `json:"fileExists"`
}

type CleanupQueryUserFile struct {
	ID        int64  `json:"id"`
	Slug      string `json:"slug"`
	UserID    int64  `json:"userId"`
	Username  string `json:"username"`
	FileName  string `json:"fileName"`
	FileSize  int64  `json:"fileSize"`
	CreatedAt int64  `json:"createdAt"`
}

type CleanupQueryResult struct {
	PhysicalFile *CleanupQueryPhysicalFile `json:"physicalFile,omitempty"`
	UserFiles    []CleanupQueryUserFile    `json:"userFiles"`
	TotalUploads int                       `json:"totalUploads"`
	UniqueUsers  int                       `json:"uniqueUsers"`
}

type DeleteActionResult struct {
	Deleted bool   `json:"deleted"`
	Message string `json:"message"`
}

func (s *AdminService) CleanupQuery(ctx context.Context, slug, hash string) (*CleanupQueryResult, error) {
	result := &CleanupQueryResult{UserFiles: []CleanupQueryUserFile{}}

	// Query by slug → find user_files, then physical_file
	if slug != "" {
		rows, err := s.pg.Query(ctx, `
			SELECT uf.id, uf.slug, uf.user_id, u.username, uf.file_name, uf.file_size,
			       uf.physical_file_id, EXTRACT(EPOCH FROM uf.created_at)::bigint
			FROM user_files uf
			JOIN users u ON u.id = uf.user_id
			WHERE uf.slug = $1
		`, slug)
		if err != nil {
			return nil, fmt.Errorf("query by slug: %w", err)
		}
		defer rows.Close()

		var pfID *int64
		for rows.Next() {
			var uf CleanupQueryUserFile
			var pID pgtype.Int8
			var createdAt int64
			if err := rows.Scan(&uf.ID, &uf.Slug, &uf.UserID, &uf.Username, &uf.FileName, &uf.FileSize, &pID, &createdAt); err != nil {
				return nil, fmt.Errorf("scan user file: %w", err)
			}
			uf.CreatedAt = createdAt
			if pID.Valid {
				pfID = &pID.Int64
			}
			result.UserFiles = append(result.UserFiles, uf)
		}
		if err := rows.Err(); err != nil {
			return nil, err
		}

		if pfID != nil {
			pf, err := s.queries.GetPhysicalFileByID(ctx, *pfID)
			if err != nil && !errors.Is(err, pgx.ErrNoRows) {
				return nil, fmt.Errorf("get physical file: %w", err)
			}
			if err == nil {
				result.PhysicalFile = cleanupPhysicalFileFromRow(pf, s.storageRoot, s.filesDir)
			}
		}
	}

	// Query by hash → find physical_file, then user_files
	if hash != "" {
		pf, err := s.queries.GetPhysicalFileByHash(ctx, sqlc.GetPhysicalFileByHashParams{
			FileHash: hash,
			HashAlgo: "sha256",
		})
		if err != nil && !errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("get physical file by hash: %w", err)
		}
		if err != nil {
			// Try hash prefix match
			row := s.pg.QueryRow(ctx, `
				SELECT id, file_hash, file_size, storage_path, mime_type
				FROM physical_files WHERE file_hash LIKE $1 || '%' LIMIT 1
			`, hash)
			var pid int64
			var fh, sp, mt string
			var fs int64
			if err := row.Scan(&pid, &fh, &fs, &sp, &mt); err != nil {
				return result, nil
			}
			pf = sqlc.PhysicalFile{ID: pid, FileHash: fh, FileSize: fs, StoragePath: sp, MimeType: mt}
		}

		physical := cleanupPhysicalFileFromRow(pf, s.storageRoot, s.filesDir)
		result.PhysicalFile = physical

		rows, err := s.pg.Query(ctx, `
			SELECT uf.id, uf.slug, uf.user_id, u.username, uf.file_name, uf.file_size,
			       EXTRACT(EPOCH FROM uf.created_at)::bigint
			FROM user_files uf
			JOIN users u ON u.id = uf.user_id
			WHERE uf.physical_file_id = $1
		`, pf.ID)
		if err != nil {
			return nil, fmt.Errorf("query user files by physical id: %w", err)
		}
		defer rows.Close()
		for rows.Next() {
			var uf CleanupQueryUserFile
			var createdAt int64
			if err := rows.Scan(&uf.ID, &uf.Slug, &uf.UserID, &uf.Username, &uf.FileName, &uf.FileSize, &createdAt); err != nil {
				return nil, fmt.Errorf("scan user file: %w", err)
			}
			uf.CreatedAt = createdAt
			result.UserFiles = append(result.UserFiles, uf)
		}
		if err := rows.Err(); err != nil {
			return nil, err
		}
	}

	seen := make(map[int64]bool)
	for _, uf := range result.UserFiles {
		result.TotalUploads++
		if !seen[uf.UserID] {
			seen[uf.UserID] = true
			result.UniqueUsers++
		}
	}
	return result, nil
}

func cleanupPhysicalFileFromRow(pf sqlc.PhysicalFile, storageRoot, filesDir string) *CleanupQueryPhysicalFile {
	exists := true
	absPath := storage.AbsPath(storageRoot, pf.FileHash, filesDir)
	if _, err := os.Stat(absPath); os.IsNotExist(err) {
		exists = false
	}
	return &CleanupQueryPhysicalFile{
		ID:          pf.ID,
		FileHash:    pf.FileHash,
		FileSize:    pf.FileSize,
		StoragePath: pf.StoragePath,
		MimeType:    pf.MimeType,
		FileExists:  exists,
	}
}

func (s *AdminService) CleanupDeleteUserFile(ctx context.Context, userFileID int64) (*DeleteActionResult, error) {
	row := s.pg.QueryRow(ctx, `
		SELECT uf.id, uf.user_id, uf.file_size, uf.physical_file_id
		FROM user_files uf WHERE uf.id = $1
	`, userFileID)
	var id, userID, fileSize int64
	var pfID pgtype.Int8
	if err := row.Scan(&id, &userID, &fileSize, &pfID); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, model.ErrNotFound
		}
		return nil, fmt.Errorf("get user file: %w", err)
	}

	tx, err := s.pg.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback(ctx)

	qtx := s.queries.WithTx(tx)

	if pfID.Valid {
		_, err := qtx.AtomicIncrementStorage(ctx, sqlc.AtomicIncrementStorageParams{
			UserID:      userID,
			StorageUsed: -fileSize,
		})
		if err != nil && !errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("decrement storage: %w", err)
		}
	}

	if err := qtx.DeleteFile(ctx, id); err != nil {
		return nil, fmt.Errorf("delete user file: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("commit tx: %w", err)
	}

	return &DeleteActionResult{Deleted: true, Message: "User file deleted"}, nil
}

func (s *AdminService) CleanupDeletePhysicalFile(ctx context.Context, physicalFileID int64) (*DeleteActionResult, error) {
	pf, err := s.queries.GetPhysicalFileByID(ctx, physicalFileID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, model.ErrNotFound
		}
		return nil, fmt.Errorf("get physical file: %w", err)
	}

	tx, err := s.pg.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback(ctx)

	qtx := s.queries.WithTx(tx)

	// Delete all user_files referencing this physical file
	ufRows, err := s.pg.Query(ctx, `
		SELECT uf.id, uf.user_id, uf.file_size FROM user_files uf WHERE uf.physical_file_id = $1
	`, physicalFileID)
	if err != nil {
		return nil, fmt.Errorf("query user files: %w", err)
	}
	var userFileIDs []int64
	for ufRows.Next() {
		var id, uid, fs int64
		if err := ufRows.Scan(&id, &uid, &fs); err != nil {
			ufRows.Close()
			return nil, fmt.Errorf("scan user file: %w", err)
		}
		userFileIDs = append(userFileIDs, id)
		_, err := qtx.AtomicIncrementStorage(ctx, sqlc.AtomicIncrementStorageParams{
			UserID:      uid,
			StorageUsed: -fs,
		})
		if err != nil && !errors.Is(err, pgx.ErrNoRows) {
			ufRows.Close()
			return nil, fmt.Errorf("decrement storage: %w", err)
		}
	}
	ufRows.Close()

	for _, id := range userFileIDs {
		if err := qtx.DeleteFile(ctx, id); err != nil {
			return nil, fmt.Errorf("delete user file %d: %w", id, err)
		}
	}

	// Delete media_items referencing this physical file
	if _, err := s.pg.Exec(ctx, `DELETE FROM media_items WHERE physical_file_id = $1`, physicalFileID); err != nil {
		return nil, fmt.Errorf("delete media items: %w", err)
	}

	// Delete from disk
	absPath := storage.AbsPath(s.storageRoot, pf.FileHash, s.filesDir)
	if err := os.Remove(absPath); err != nil && !os.IsNotExist(err) {
		s.logger.Warn().Err(err).Str("path", absPath).Msg("failed to remove physical file from disk")
	}

	if err := qtx.DeletePhysicalFile(ctx, physicalFileID); err != nil {
		return nil, fmt.Errorf("delete physical file: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("commit tx: %w", err)
	}

	return &DeleteActionResult{
		Deleted: true,
		Message: fmt.Sprintf("Physical file deleted, %d user records cleaned up", len(userFileIDs)),
	}, nil
}

type AdminActivityLogItem struct {
	ID           int64           `json:"id"`
	UserID       int64           `json:"userId"`
	Username     string          `json:"username"`
	Action       string          `json:"action"`
	ActionLabel  string          `json:"actionLabel"`
	ResourceType string          `json:"resourceType"`
	ResourceName string          `json:"resourceName"`
	IP           string          `json:"ip"`
	IPRegion     string          `json:"ipRegion"`
	UserAgent    string          `json:"userAgent"`
	OS           string          `json:"os"`
	Browser      string          `json:"browser"`
	Extra        json.RawMessage `json:"extra"`
	CreatedAt    time.Time       `json:"createdAt"`
}

type AdminListActivityLogsParams struct {
	Limit       int
	Offset      int
	UserID      *int64
	Action      string
	IP          string
	CreatedFrom *time.Time
	CreatedTo   *time.Time
	Locale      string
}

func (s *AdminService) ListActivityLogs(ctx context.Context, params AdminListActivityLogsParams) ([]AdminActivityLogItem, int, error) {
	where := "TRUE"
	args := []any{}
	argIdx := 1

	if params.UserID != nil {
		where += fmt.Sprintf(" AND l.user_id = $%d", argIdx)
		args = append(args, *params.UserID)
		argIdx++
	}
	if params.Action != "" {
		where += fmt.Sprintf(" AND l.action = $%d", argIdx)
		args = append(args, params.Action)
		argIdx++
	}
	if params.IP != "" {
		where += fmt.Sprintf(" AND l.ip = $%d", argIdx)
		args = append(args, params.IP)
		argIdx++
	}
	if params.CreatedFrom != nil {
		where += fmt.Sprintf(" AND l.created_at >= $%d", argIdx)
		args = append(args, *params.CreatedFrom)
		argIdx++
	}
	if params.CreatedTo != nil {
		where += fmt.Sprintf(" AND l.created_at <= $%d", argIdx)
		args = append(args, *params.CreatedTo)
		argIdx++
	}

	var total int
	countQ := "SELECT COUNT(*) FROM user_activity_logs l WHERE " + where
	if err := s.pg.QueryRow(ctx, countQ, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("count activity logs: %w", err)
	}

	q := fmt.Sprintf(`
		SELECT l.id, l.user_id, COALESCE(u.username, ''),
		       l.action, COALESCE(l.resource_type, ''), COALESCE(l.resource_name, ''),
		       COALESCE(l.ip, ''), COALESCE(l.ip_region, ''), COALESCE(l.user_agent, ''),
		       COALESCE(l.os, ''), COALESCE(l.browser, ''), l.extra, l.created_at
		FROM user_activity_logs l
		LEFT JOIN users u ON u.id = l.user_id
		WHERE %s
		ORDER BY l.created_at DESC
		LIMIT $%d OFFSET $%d
	`, where, argIdx, argIdx+1)

	args = append(args, params.Limit, params.Offset)
	rows, err := s.pg.Query(ctx, q, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("list activity logs: %w", err)
	}
	defer rows.Close()

	var items []AdminActivityLogItem
	for rows.Next() {
		var item AdminActivityLogItem
		if err := rows.Scan(
			&item.ID, &item.UserID, &item.Username,
			&item.Action, &item.ResourceType, &item.ResourceName,
			&item.IP, &item.IPRegion, &item.UserAgent,
			&item.OS, &item.Browser, &item.Extra, &item.CreatedAt,
		); err != nil {
			return nil, 0, fmt.Errorf("scan activity log: %w", err)
		}
		item.ActionLabel = i18n.ActionLabel(item.Action, params.Locale)
		items = append(items, item)
	}
	return items, total, nil
}

type AdminActionLabelItem struct {
	Action string `json:"action"`
	Label  string `json:"label"`
}

func (s *AdminService) ListActivityLogActions(ctx context.Context, locale string) ([]AdminActionLabelItem, error) {
	rows, err := s.pg.Query(ctx, `SELECT DISTINCT action FROM user_activity_logs ORDER BY action`)
	if err != nil {
		return nil, fmt.Errorf("list actions: %w", err)
	}
	defer rows.Close()

	var items []AdminActionLabelItem
	for rows.Next() {
		var action string
		if err := rows.Scan(&action); err != nil {
			return nil, fmt.Errorf("scan action: %w", err)
		}
		items = append(items, AdminActionLabelItem{
			Action: action,
			Label:  i18n.ActionLabel(action, locale),
		})
	}
	return items, nil
}
