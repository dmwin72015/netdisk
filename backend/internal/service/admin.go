package service

import (
	"context"
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
	Limit        int
	Offset       int
	Search       string
	Role         string
	Sort         string
	CreatedAfter *time.Time
	CreatedBefore *time.Time
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


type CleanupFileResult struct {
	Slug          string                `json:"slug"`
	UploadTasks   []CleanupUploadTask   `json:"uploadTasks"`
	UserFiles     []CleanupUserFile     `json:"userFiles"`
	PhysicalFiles []CleanupPhysicalFile `json:"physicalFiles"`
	Deleted       bool                  `json:"deleted"`
	Message       string                `json:"message,omitempty"`
}

type CleanupUploadTask struct {
	ID           int64  `json:"id"`
	OwnerUserID  int64  `json:"ownerUserId"`
	Username     string `json:"username,omitempty"`
	Status       string `json:"status"`
	FileSize     int64  `json:"fileSize"`
	OriginalName string `json:"originalName"`
	CreatedAt    int64  `json:"createdAt"`
}

type CleanupUserFile struct {
	ID               int64   `json:"id"`
	UserID           int64   `json:"userId"`
	Username         string  `json:"username"`
	FileName         string  `json:"fileName"`
	FileSize         int64   `json:"fileSize"`
	PhysicalFileID   *int64  `json:"physicalFileId,omitempty"`
	CreatedAt        int64   `json:"createdAt"`
}

type CleanupPhysicalFile struct {
	ID          int64   `json:"id"`
	FileHash    string  `json:"fileHash"`
	FileSize    int64   `json:"fileSize"`
	StoragePath string  `json:"storagePath"`
	RefCount    *int64  `json:"refCount"`
}

func (s *AdminService) CleanupFile(ctx context.Context, slug string, doDelete bool) (*CleanupFileResult, error) {
	result := &CleanupFileResult{
		Slug:          slug,
		UploadTasks:   []CleanupUploadTask{},
		UserFiles:     []CleanupUserFile{},
		PhysicalFiles: []CleanupPhysicalFile{},
	}

	task, err := s.queries.GetUploadTaskBySlug(ctx, slug)
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return nil, fmt.Errorf("query upload task: %w", err)
	}
	if err == nil {
		var username string
		_ = s.pg.QueryRow(ctx, `SELECT username FROM users WHERE id = $1`, task.OwnerUserID).Scan(&username)

		var createdAt int64
		if task.CreatedAt.Valid {
			createdAt = task.CreatedAt.Time.Unix()
		}

		result.UploadTasks = append(result.UploadTasks, CleanupUploadTask{
			ID:           task.ID,
			OwnerUserID:  task.OwnerUserID,
			Username:     username,
			Status:       task.Status,
			FileSize:     task.FileSize,
			OriginalName: task.OriginalName,
			CreatedAt:    createdAt,
		})
	}

	rows, err := s.pg.Query(ctx, `
		SELECT uf.id, uf.user_id, u.username, uf.file_name, uf.file_size,
		       uf.physical_file_id, EXTRACT(EPOCH FROM uf.created_at)::bigint
		FROM user_files uf
		JOIN users u ON u.id = uf.user_id
		WHERE uf.slug = $1
	`, slug)
	if err != nil {
		return nil, fmt.Errorf("query user files: %w", err)
	}
	defer rows.Close()

	pfIDSet := make(map[int64]bool)
	for rows.Next() {
		var uf CleanupUserFile
		var pfID pgtype.Int8
		var createdAt int64
		if err := rows.Scan(&uf.ID, &uf.UserID, &uf.Username, &uf.FileName, &uf.FileSize, &pfID, &createdAt); err != nil {
			return nil, fmt.Errorf("scan user file: %w", err)
		}
		if pfID.Valid {
			id := pfID.Int64
			uf.PhysicalFileID = &id
			pfIDSet[id] = true
		}
		uf.CreatedAt = createdAt
		result.UserFiles = append(result.UserFiles, uf)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	if len(pfIDSet) > 0 {
		ids := make([]int64, 0, len(pfIDSet))
		for id := range pfIDSet {
			ids = append(ids, id)
		}

		pfRows, err := s.pg.Query(ctx, `
			SELECT id, file_hash, file_size, storage_path
			FROM physical_files
			WHERE id = ANY($1::bigint[])
		`, ids)
		if err != nil {
			return nil, fmt.Errorf("query physical files: %w", err)
		}
		defer pfRows.Close()

		pfPathMap := make(map[int64]string)
		for pfRows.Next() {
			var pf CleanupPhysicalFile
			if err := pfRows.Scan(&pf.ID, &pf.FileHash, &pf.FileSize, &pf.StoragePath); err != nil {
				return nil, fmt.Errorf("scan physical file: %w", err)
			}
			refCount, err := s.queries.CountReferencesByFileID(ctx, pgtype.Int8{Int64: pf.ID, Valid: true})
			if err != nil {
				return nil, fmt.Errorf("count references: %w", err)
			}
			v := int64(refCount)
			pf.RefCount = &v
			pfPathMap[pf.ID] = pf.StoragePath
			result.PhysicalFiles = append(result.PhysicalFiles, pf)
		}
		if err := pfRows.Err(); err != nil {
			return nil, err
		}
	}

	// If no results found by slug, try searching by physical file hash prefix
	if len(result.UploadTasks) == 0 && len(result.UserFiles) == 0 {
		pfRow := s.pg.QueryRow(ctx, `
			SELECT id, file_size, storage_path FROM physical_files
			WHERE file_hash LIKE $1 || '%'
			LIMIT 1
		`, slug)
		var pfID int64
		var pfSize int64
		var pfPath string
		if err := pfRow.Scan(&pfID, &pfSize, &pfPath); err == nil {
			pfRows, err := s.pg.Query(ctx, `
				SELECT uf.id, uf.user_id, u.username, uf.file_name, uf.file_size,
				       uf.physical_file_id, EXTRACT(EPOCH FROM uf.created_at)::bigint
				FROM user_files uf
				JOIN users u ON u.id = uf.user_id
				WHERE uf.physical_file_id = $1
			`, pfID)
			if err == nil {
				for pfRows.Next() {
					var uf CleanupUserFile
					var physicalFileID pgtype.Int8
					var createdAt int64
					if err := pfRows.Scan(&uf.ID, &uf.UserID, &uf.Username, &uf.FileName, &uf.FileSize, &physicalFileID, &createdAt); err == nil {
						if physicalFileID.Valid {
							id := physicalFileID.Int64
							uf.PhysicalFileID = &id
						}
						uf.CreatedAt = createdAt
						result.UserFiles = append(result.UserFiles, uf)
					}
				}
				pfRows.Close()
			}
			if len(result.UserFiles) > 0 {
				refCount, _ := s.queries.CountReferencesByFileID(ctx, pgtype.Int8{Int64: pfID, Valid: true})
				rc := int64(refCount)
				result.PhysicalFiles = append(result.PhysicalFiles, CleanupPhysicalFile{
					ID:          pfID,
					FileHash:    slug,
					FileSize:    pfSize,
					StoragePath: pfPath,
					RefCount:    &rc,
				})
				for i := range result.UserFiles {
					if result.UserFiles[i].PhysicalFileID != nil {
					}
				}
			}
		}
	}

	if !doDelete {
		return result, nil
	}

	tx, err := s.pg.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback(ctx)

	qtx := s.queries.WithTx(tx)

	if len(result.UploadTasks) > 0 {
		if err := qtx.DeleteUploadTaskBySlug(ctx, sqlc.DeleteUploadTaskBySlugParams{Slug: slug, OwnerUserID: result.UploadTasks[0].OwnerUserID}); err != nil {
			return nil, fmt.Errorf("delete upload task: %w", err)
		}
	}

	for _, uf := range result.UserFiles {
		if uf.PhysicalFileID != nil {
			_, err := qtx.AtomicIncrementStorage(ctx, sqlc.AtomicIncrementStorageParams{
				UserID:      uf.UserID,
				StorageUsed: -uf.FileSize,
			})
			if err != nil && !errors.Is(err, pgx.ErrNoRows) {
				return nil, fmt.Errorf("decrement storage: %w", err)
			}
		}
		if err := qtx.DeleteFile(ctx, uf.ID); err != nil {
			return nil, fmt.Errorf("delete user file: %w", err)
		}
	}

	for _, pf := range result.PhysicalFiles {
		if *pf.RefCount == 0 {
			_ = os.Remove(storage.AbsPath(s.storageRoot, pf.FileHash, s.filesDir))
			if err := qtx.DeletePhysicalFile(ctx, pf.ID); err != nil {
				return nil, fmt.Errorf("delete physical file: %w", err)
			}
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("commit tx: %w", err)
	}

	result.Deleted = true
	result.Message = "Cleanup completed"
	return result, nil
}
