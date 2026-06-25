package service

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	gonanoid "github.com/matoous/go-nanoid/v2"

	"github.com/netdisk/server/internal/config"
	"github.com/netdisk/server/internal/db"
	"github.com/netdisk/server/internal/db/sqlc"
	"github.com/netdisk/server/internal/model"
	"github.com/netdisk/server/internal/storage"
	"github.com/netdisk/server/pkg/fileutil"
	"golang.org/x/crypto/bcrypt"
)

type FilesService struct {
	queries *sqlc.Queries
	pg      *pgxpool.Pool
	cfg     *config.Config
	store   *storage.Local
}

func NewFilesService(queries *sqlc.Queries, pg *pgxpool.Pool, cfg *config.Config, store *storage.Local) *FilesService {
	return &FilesService{queries: queries, pg: pg, cfg: cfg, store: store}
}

const SystemDirMediaUploads = "media_uploads"

type FileItem struct {
	Slug         string  `json:"slug"`
	FileName     string  `json:"fileName"`
	IsDir        bool    `json:"isDir"`
	HasPassword  bool    `json:"hasPassword"`
	FileSize     int64   `json:"fileSize"`
	MimeType     *string `json:"mimeType"`
	FileCategory string  `json:"fileCategory"`
	IsStarred    bool    `json:"isStarred"`
	IsSystem     bool    `json:"isSystem"`
	SystemKind   *string `json:"systemKind,omitempty"`
	ParentSlug   *string `json:"parentSlug,omitempty"`
	ParentName   *string `json:"parentName,omitempty"`
	CreatedAt    string  `json:"createdAt"`
	UpdatedAt    string  `json:"updatedAt"`
	HashAlgo     *string `json:"hashAlgo,omitempty"`
	FileHash     *string `json:"fileHash,omitempty"`
}

type SystemDirOptions struct {
	Kind       string
	Name       string
	ParentSlug string
}

type ConflictResponse struct {
	Status   string    `json:"status"`
	Message  string    `json:"message,omitempty"`
	Existing *FileItem `json:"existing,omitempty"`
}

type ImportResponse struct {
	FileSlug string `json:"fileSlug"`
	FileName string `json:"fileName"`
}

func (s *FilesService) SetDirectoryLock(ctx context.Context, userID int64, sessionID, dirSlug, password string) error {
	if dirSlug == "" || len(password) < 4 || len(password) > 128 {
		return model.ErrInvalidInput
	}
	dir, err := s.queries.GetFileBySlugForUser(ctx, sqlc.GetFileBySlugForUserParams{Slug: dirSlug, UserID: userID})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return model.ErrNotFound
		}
		return fmt.Errorf("get directory: %w", err)
	}
	if !dir.IsDir || dir.IsTrashed || dir.IsSystem {
		return model.ErrInvalidInput
	}

	// If the directory already has a lock, verify the old password before changing it.
	// This is consistent with ClearDirectoryLock and avoids the unlock-then-set two-step.
	if dir.LockPasswordHash.Valid && dir.LockPasswordHash.String != "" {
		if password == "" {
			return model.ErrInvalidInput
		}
		if err := bcrypt.CompareHashAndPassword([]byte(dir.LockPasswordHash.String), []byte(password)); err != nil {
			return model.ErrUnauthorized
		}
	} else {
		// No existing lock — only check parent directories are not locked.
		if err := ensureFileUnlocked(ctx, s.pg, userID, sessionID, dir.ID); err != nil {
			return err
		}
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), s.cfg.Limits.BcryptCost)
	if err != nil {
		return fmt.Errorf("hash directory password: %w", err)
	}
	_, err = s.pg.Exec(ctx, `
		UPDATE user_files
		SET lock_password_hash = $1, locked_at = NOW(), updated_at = NOW()
		WHERE id = $2 AND user_id = $3 AND is_dir = TRUE
	`, string(hash), dir.ID, userID)
	if err != nil {
		return fmt.Errorf("set directory lock: %w", err)
	}
	if sessionID != "" {
		if _, err := s.pg.Exec(ctx, `DELETE FROM user_directory_unlocks WHERE user_id = $1 AND directory_id = $2 AND session_id = $3`, userID, dir.ID, sessionID); err != nil {
			slog.Warn("failed to remove directory unlock after setting lock", "userID", userID, "dirID", dir.ID, "error", err)
		}
	}
	return nil
}

func (s *FilesService) ClearDirectoryLock(ctx context.Context, userID int64, sessionID, dirSlug, password string) error {
	dir, err := s.queries.GetFileBySlugForUser(ctx, sqlc.GetFileBySlugForUserParams{Slug: dirSlug, UserID: userID})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return model.ErrNotFound
		}
		return fmt.Errorf("get directory: %w", err)
	}
	if !dir.IsDir || dir.IsTrashed || dir.IsSystem {
		return model.ErrInvalidInput
	}
	locked, err := findLockedAncestor(ctx, s.pg, userID, sessionID, dir.ID)
	if err != nil {
		return err
	}
	if locked != nil && locked.ID != dir.ID {
		return model.ErrDirectoryLocked
	}
	if dir.LockPasswordHash.Valid && dir.LockPasswordHash.String != "" {
		if password == "" {
			return model.ErrInvalidInput
		}
		if err := bcrypt.CompareHashAndPassword([]byte(dir.LockPasswordHash.String), []byte(password)); err != nil {
			return model.ErrUnauthorized
		}
	}
	_, err = s.pg.Exec(ctx, `
		UPDATE user_files
		SET lock_password_hash = NULL, locked_at = NULL, updated_at = NOW()
		WHERE id = $1 AND user_id = $2
	`, dir.ID, userID)
	if err != nil {
		return fmt.Errorf("clear directory lock: %w", err)
	}
	_, _ = s.pg.Exec(ctx, `DELETE FROM user_directory_unlocks WHERE user_id = $1 AND directory_id = $2`, userID, dir.ID)
	return nil
}

func (s *FilesService) UnlockDirectory(ctx context.Context, userID int64, sessionID, dirSlug, password string, ttlHours int) error {
	if sessionID == "" || dirSlug == "" || password == "" {
		return model.ErrInvalidInput
	}
	dir, err := s.queries.GetFileBySlugForUser(ctx, sqlc.GetFileBySlugForUserParams{Slug: dirSlug, UserID: userID})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return model.ErrNotFound
		}
		return fmt.Errorf("get directory: %w", err)
	}
	if !dir.IsDir || dir.IsTrashed {
		return model.ErrInvalidInput
	}
	lockDir := dir
	var lockedDirID int64
	var passwordHash string
	if lockDir.LockPasswordHash.Valid && lockDir.LockPasswordHash.String != "" {
		lockedDirID = lockDir.ID
		passwordHash = lockDir.LockPasswordHash.String
	} else {
		locked, err := findLockedAncestor(ctx, s.pg, userID, sessionID, dir.ID)
		if err != nil {
			return err
		}
		if locked == nil {
			return nil
		}
		lockedDirID = locked.ID
		passwordHash = locked.LockPasswordHash
	}
	if passwordHash == "" {
		return model.ErrInvalidInput
	}
	if err := bcrypt.CompareHashAndPassword([]byte(passwordHash), []byte(password)); err != nil {
		return model.ErrUnauthorized
	}

	expiresAt, permanent := directoryUnlockExpiry(ttlHours)
	if permanent {
		_, err = s.pg.Exec(ctx, `
			INSERT INTO user_directory_unlocks (user_id, directory_id, session_id, expires_at, updated_at)
			VALUES ($1, $2, $3, NULL, NOW())
			ON CONFLICT (user_id, directory_id, session_id)
			DO UPDATE SET expires_at = NULL, updated_at = NOW()
		`, userID, lockedDirID, sessionID)
	} else {
		_, err = s.pg.Exec(ctx, `
			INSERT INTO user_directory_unlocks (user_id, directory_id, session_id, expires_at, updated_at)
			VALUES ($1, $2, $3, $4, NOW())
			ON CONFLICT (user_id, directory_id, session_id)
			DO UPDATE SET expires_at = EXCLUDED.expires_at, updated_at = NOW()
		`, userID, lockedDirID, sessionID, expiresAt)
	}
	if err != nil {
		return fmt.Errorf("store directory unlock: %w", err)
	}
	return nil
}

func directoryUnlockExpiry(ttlHours int) (time.Time, bool) {
	if ttlHours == PermanentDirectoryUnlockTTL {
		return time.Time{}, true
	}
	switch ttlHours {
	case 1, 2, 6, 24:
	default:
		ttlHours = DefaultDirectoryUnlockTTLHours
	}
	return time.Now().Add(time.Duration(ttlHours) * time.Hour), false
}

func (s *FilesService) ResolveParent(ctx context.Context, userID int64, parentSlug string) (sqlc.UserFile, error) {
	parent, err := s.queries.GetFileBySlugForUser(ctx, sqlc.GetFileBySlugForUserParams{
		Slug:   parentSlug,
		UserID: userID,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return sqlc.UserFile{}, model.ErrNotFound
		}
		return sqlc.UserFile{}, fmt.Errorf("get parent: %w", err)
	}
	if parent.IsTrashed {
		return sqlc.UserFile{}, model.ErrNotFound
	}
	if !parent.IsDir {
		return sqlc.UserFile{}, model.ErrNotFound
	}
	return parent, nil
}

// ListUserFiles is the unified file listing method powered by Squirrel.
func (s *FilesService) ListUserFiles(ctx context.Context, params db.ListFilesParams, sessionID ...string) ([]FileItem, int, error) {
	sid := ""
	if len(sessionID) > 0 {
		sid = sessionID[0]
	}
	if !params.IgnoreParentID && params.ParentID != nil {
		if err := ensureFileUnlocked(ctx, s.pg, params.UserID, sid, *params.ParentID); err != nil {
			return nil, 0, err
		}
	}

	sql, args, countSql, countArgs, err := db.BuildListFilesQuery(params)
	if err != nil {
		return nil, 0, err
	}

	var total int
	err = s.pg.QueryRow(ctx, countSql, countArgs...).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("count files: %w", err)
	}

	rows, err := s.pg.Query(ctx, sql, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("list files: %w", err)
	}
	defer rows.Close()

	fileRows, err := db.ScanFileRows(rows)
	if err != nil {
		return nil, 0, fmt.Errorf("scan files: %w", err)
	}
	if params.IgnoreParentID || params.IsTrashed {
		before := len(fileRows)
		fileRows, err = s.filterVisibleRows(ctx, params.UserID, sid, fileRows)
		if err != nil {
			return nil, 0, err
		}
		// Adjust total to account for rows filtered out by directory locks.
		// This is an approximation — we subtract only the rows filtered from
		// the current page, which keeps the count closer to reality than the
		// raw SQL COUNT that includes locked files.
		if before > 0 && len(fileRows) < before {
			total -= before - len(fileRows)
			if total < 0 {
				total = 0
			}
		}
	}

	return fileRowsToItems(fileRows), total, nil
}

func (s *FilesService) filterVisibleRows(ctx context.Context, userID int64, sessionID string, rows []db.FileRow) ([]db.FileRow, error) {
	if len(rows) == 0 {
		return rows, nil
	}

	// Collect file IDs for batch lock filtering.
	fileIDs := make([]int64, len(rows))
	for i, row := range rows {
		fileIDs[i] = row.ID
	}

	visibleIDs, err := filterLockedFileIDs(ctx, s.pg, userID, sessionID, fileIDs)
	if err != nil {
		return nil, err
	}
	visibleSet := make(map[int64]bool, len(visibleIDs))
	for _, id := range visibleIDs {
		visibleSet[id] = true
	}

	// Self-locked directories are already visible (filterLockedFileIDs only hides
	// files whose ancestor is locked, not files that are themselves locked).
	visible := rows[:0]
	for _, row := range rows {
		if visibleSet[row.ID] {
			visible = append(visible, row)
		}
	}
	return visible, nil
}

func (s *FilesService) ListRecentFiles(ctx context.Context, userID int64, sessionID string, limit int) ([]FileItem, int, error) {
	params := db.ListFilesParams{
		UserID:         userID,
		IncludeDirs:    false,
		IgnoreParentID: true, // show files from all directories
		SortBy:         "created_at",
		SortDir:        "DESC",
		Page:           1,
		PageSize:       limit,
	}

	return s.ListUserFiles(ctx, params, sessionID)
}

func (s *FilesService) EnsureSystemDir(ctx context.Context, userID int64, opts SystemDirOptions) (*FileItem, error) {
	if opts.Kind == "" || opts.Name == "" || len(opts.Name) > 100 {
		return nil, model.ErrInvalidInput
	}

	parentID, err := s.resolveParentID(ctx, userID, opts.ParentSlug)
	if err != nil {
		return nil, err
	}

	existing, err := s.queries.GetSystemDirByKind(ctx, sqlc.GetSystemDirByKindParams{
		UserID:     userID,
		ParentID:   parentID,
		SystemKind: pgtype.Text{String: opts.Kind, Valid: true},
	})
	if err == nil {
		item := fileToItem(existing)
		return &item, nil
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		return nil, fmt.Errorf("get system dir: %w", err)
	}

	conflict, err := s.queries.CheckNameConflict(ctx, sqlc.CheckNameConflictParams{
		UserID:   userID,
		ParentID: parentID,
		FileName: opts.Name,
		IsSystem: true,
	})
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return nil, fmt.Errorf("check system dir conflict: %w", err)
	}
	if conflict.ID != 0 {
		return nil, model.ErrNameConflict
	}

	slug, err := gonanoid.New(21)
	if err != nil {
		return nil, fmt.Errorf("generate slug: %w", err)
	}

	f, err := s.queries.CreateFile(ctx, sqlc.CreateFileParams{
		Slug:         slug,
		UserID:       userID,
		ParentID:     parentID,
		ParentSlug:   pgtype.Text{String: opts.ParentSlug, Valid: opts.ParentSlug != ""},
		FileName:     opts.Name,
		IsDir:        true,
		FileSize:     0,
		FileCategory: string(fileutil.CategoryFolder),
		IsSystem:     true,
		SystemKind:   pgtype.Text{String: opts.Kind, Valid: true},
	})
	if err != nil {
		if isUniqueViolation(err) {
			existing, getErr := s.queries.GetSystemDirByKind(ctx, sqlc.GetSystemDirByKindParams{
				UserID:     userID,
				ParentID:   parentID,
				SystemKind: pgtype.Text{String: opts.Kind, Valid: true},
			})
			if getErr == nil {
				item := fileToItem(existing)
				return &item, nil
			}
			return nil, model.ErrNameConflict
		}
		return nil, fmt.Errorf("create system dir: %w", err)
	}

	item := fileToItem(f)
	return &item, nil
}

type BreadcrumbItem struct {
	Slug     string `json:"slug"`
	FileName string `json:"fileName"`
}

func (s *FilesService) GetBreadcrumb(ctx context.Context, userID int64, sessionID, slug string) ([]BreadcrumbItem, error) {
	f, err := s.queries.GetFileBySlugForUser(ctx, sqlc.GetFileBySlugForUserParams{
		Slug:   slug,
		UserID: userID,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, model.ErrNotFound
		}
		return nil, fmt.Errorf("get file: %w", err)
	}
	if err := ensureFileUnlocked(ctx, s.pg, userID, sessionID, f.ID); err != nil {
		return nil, err
	}

	ancestors, err := s.queries.GetAncestors(ctx, f.ID)
	if err != nil {
		return nil, fmt.Errorf("get ancestors: %w", err)
	}

	result := make([]BreadcrumbItem, 0, len(ancestors))
	for _, a := range ancestors {
		result = append(result, BreadcrumbItem{
			Slug:     a.Slug,
			FileName: a.FileName,
		})
	}
	return result, nil
}

func (s *FilesService) Mkdir(ctx context.Context, userID int64, sessionID, dirName, parentSlug string) (*FileItem, error) {
	if dirName == "" || len(dirName) > 100 {
		return nil, model.ErrInvalidInput
	}

	parentID, err := s.resolveParentID(ctx, userID, parentSlug)
	if err != nil {
		return nil, err
	}
	if err := s.ensureResolvedParentUnlocked(ctx, userID, sessionID, parentID); err != nil {
		return nil, err
	}

	conflict, err := s.queries.CheckNameConflict(ctx, sqlc.CheckNameConflictParams{
		UserID:   userID,
		ParentID: parentID,
		FileName: dirName,
		IsSystem: false,
	})
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return nil, fmt.Errorf("check conflict: %w", err)
	}
	if conflict.ID != 0 {
		return nil, model.ErrNameConflict
	}

	slug, err := gonanoid.New(21)
	if err != nil {
		return nil, fmt.Errorf("generate slug: %w", err)
	}

	f, err := s.queries.CreateFile(ctx, sqlc.CreateFileParams{
		Slug:         slug,
		UserID:       userID,
		ParentID:     parentID,
		ParentSlug:   pgtype.Text{String: parentSlug, Valid: parentSlug != ""},
		FileName:     dirName,
		IsDir:        true,
		FileSize:     0,
		FileCategory: string(fileutil.CategoryFolder),
		IsSystem:     false,
	})
	if err != nil {
		if isUniqueViolation(err) {
			return nil, model.ErrNameConflict
		}
		return nil, fmt.Errorf("create dir: %w", err)
	}

	return &FileItem{
		Slug:         f.Slug,
		FileName:     f.FileName,
		IsDir:        true,
		FileSize:     0,
		FileCategory: string(fileutil.CategoryFolder),
		CreatedAt:    f.CreatedAt.Time.Format("2006-01-02T15:04:05Z"),
		UpdatedAt:    f.UpdatedAt.Time.Format("2006-01-02T15:04:05Z"),
	}, nil
}

func (s *FilesService) CheckConflict(ctx context.Context, userID int64, sessionID, fileName, preHash, parentSlug string, fileSize int64) (*ConflictResponse, error) {
	parentID, err := s.resolveParentID(ctx, userID, parentSlug)
	if err != nil {
		return nil, err
	}
	if err := s.ensureResolvedParentUnlocked(ctx, userID, sessionID, parentID); err != nil {
		return nil, err
	}

	conflict, err := s.queries.CheckNameConflict(ctx, sqlc.CheckNameConflictParams{
		UserID:   userID,
		ParentID: parentID,
		FileName: fileName,
		IsSystem: false,
	})
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return nil, fmt.Errorf("check conflict: %w", err)
	}
	if conflict.ID != 0 {
		item := &FileItem{
			Slug:     conflict.Slug,
			FileName: conflict.FileName,
			IsDir:    conflict.IsDir,
			FileSize: conflict.FileSize,
		}
		return &ConflictResponse{
			Status:   "NAME_CONFLICT",
			Message:  "当前目录已存在同名条目，是否继续上传？",
			Existing: item,
		}, nil
	}

	if preHash != "" {
		pf, err := s.queries.GetPhysicalFileByPreHash(ctx, sqlc.GetPhysicalFileByPreHashParams{
			PreHash:  preHash,
			FileSize: fileSize,
		})
		if err == nil && pf.ID != 0 {
			return &ConflictResponse{
				Status:  "SAME_FILE_CONFLICT",
				Message: "当前目录可能已存在相同文件，是否继续检查并上传？",
				Existing: &FileItem{
					Slug:     pf.Slug,
					FileSize: pf.FileSize,
				},
			}, nil
		}
	}

	return &ConflictResponse{Status: "OK"}, nil
}

func (s *FilesService) CheckDuplicate(ctx context.Context, userID int64, sessionID, fileHash, parentSlug string) (*ConflictResponse, error) {
	parentID, err := s.resolveParentID(ctx, userID, parentSlug)
	if err != nil {
		return nil, err
	}
	if err := s.ensureResolvedParentUnlocked(ctx, userID, sessionID, parentID); err != nil {
		return nil, err
	}

	pf, err := s.queries.GetPhysicalFileByHash(ctx, sqlc.GetPhysicalFileByHashParams{
		HashAlgo: "sha256",
		FileHash: fileHash,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return &ConflictResponse{Status: "OK"}, nil
		}
		return nil, fmt.Errorf("get physical file: %w", err)
	}

	return &ConflictResponse{
		Status:  "DUPLICATE_FILE",
		Message: "当前目录已存在相同文件，是否仍然上传？",
		Existing: &FileItem{
			Slug:     pf.Slug,
			FileSize: pf.FileSize,
		},
	}, nil
}

func (s *FilesService) ImportFile(ctx context.Context, userID int64, sessionID, physicalFileSlug, fileName, parentSlug string) (*ImportResponse, error) {
	if fileName == "" || physicalFileSlug == "" {
		return nil, model.ErrInvalidInput
	}

	pf, err := s.queries.GetPhysicalFileBySlug(ctx, physicalFileSlug)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, model.ErrNotFound
		}
		return nil, fmt.Errorf("get physical file: %w", err)
	}
	if pf.Status != "completed" {
		return nil, model.ErrInvalidInput
	}

	parentID, err := s.resolveParentID(ctx, userID, parentSlug)
	if err != nil {
		return nil, err
	}
	if err := s.ensureResolvedParentUnlocked(ctx, userID, sessionID, parentID); err != nil {
		return nil, err
	}

	conflict, err := s.queries.CheckNameConflict(ctx, sqlc.CheckNameConflictParams{
		UserID:   userID,
		ParentID: parentID,
		FileName: fileName,
		IsSystem: false,
	})
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return nil, fmt.Errorf("check conflict: %w", err)
	}
	if conflict.ID != 0 {
		return nil, model.ErrNameConflict
	}

	// Atomic quota increment — ErrNoRows means quota exceeded
	_, err = s.queries.AtomicIncrementStorage(ctx, sqlc.AtomicIncrementStorageParams{
		UserID:      userID,
		StorageUsed: pf.FileSize,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, model.ErrQuotaExceeded
		}
		return nil, fmt.Errorf("increment storage: %w", err)
	}

	slug, err := gonanoid.New(21)
	if err != nil {
		return nil, fmt.Errorf("generate slug: %w", err)
	}

	mimeType := pgtype.Text{}
	if pf.MimeType != "" {
		mimeType = pgtype.Text{String: pf.MimeType, Valid: true}
	}

	category := fileutil.CategorizeMime(pf.MimeType, false)

	f, err := s.queries.CreateFile(ctx, sqlc.CreateFileParams{
		Slug:           slug,
		UserID:         userID,
		PhysicalFileID: pgtype.Int8{Int64: pf.ID, Valid: true},
		ParentID:       parentID,
		ParentSlug:     pgtype.Text{String: parentSlug, Valid: parentSlug != ""},
		FileName:       fileName,
		IsDir:          false,
		FileSize:       pf.FileSize,
		MimeType:       mimeType,
		FileCategory:   string(category),
		IsSystem:       false,
	})
	if err != nil {
		return nil, fmt.Errorf("create file: %w", err)
	}

	return &ImportResponse{
		FileSlug: f.Slug,
		FileName: f.FileName,
	}, nil
}

func (s *FilesService) TrashFile(ctx context.Context, userID int64, sessionID, fileSlug string) error {
	f, err := s.queries.GetFileBySlugForUser(ctx, sqlc.GetFileBySlugForUserParams{
		Slug:   fileSlug,
		UserID: userID,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return model.ErrNotFound
		}
		return fmt.Errorf("get file: %w", err)
	}
	if f.IsTrashed {
		return nil
	}
	if f.IsSystem {
		return model.ErrSystemFileLocked
	}
	if err := ensureFileUnlocked(ctx, s.pg, userID, sessionID, f.ID); err != nil {
		return err
	}

	// Check if directory is empty before trashing
	if f.IsDir {
		var count int64
		err = s.pg.QueryRow(ctx,
			"SELECT COUNT(*) FROM user_files WHERE user_id = $1 AND parent_id = $2 AND is_trashed = FALSE",
			userID, f.ID,
		).Scan(&count)
		if err != nil {
			return fmt.Errorf("count children: %w", err)
		}
		if count > 0 {
			return model.ErrDirNotEmpty
		}
	}

	return s.queries.SetTrashed(ctx, sqlc.SetTrashedParams{
		ID:        f.ID,
		IsTrashed: true,
	})
}

func (s *FilesService) BatchTrashFiles(ctx context.Context, userID int64, sessionID string, slugs []string) error {
	if len(slugs) == 0 {
		return nil
	}

	// Query all files by slugs in a single query
	rows, err := s.pg.Query(ctx,
		`SELECT id, slug, is_dir, is_trashed, is_system FROM user_files
		 WHERE slug = ANY($1::varchar[]) AND user_id = $2`,
		slugs, userID,
	)
	if err != nil {
		return fmt.Errorf("batch trash: query files: %w", err)
	}
	defer rows.Close()

	type foundFile struct {
		ID        int64
		Slug      string
		IsDir     bool
		IsTrashed bool
		IsSystem  bool
	}
	var files []foundFile
	var notFound []string
	foundSlugs := make(map[string]bool)

	for rows.Next() {
		var f foundFile
		if err := rows.Scan(&f.ID, &f.Slug, &f.IsDir, &f.IsTrashed, &f.IsSystem); err != nil {
			return fmt.Errorf("batch trash: scan: %w", err)
		}
		files = append(files, f)
		foundSlugs[f.Slug] = true
	}
	if err := rows.Err(); err != nil {
		return err
	}

	// Check for not-found slugs
	for _, slug := range slugs {
		if !foundSlugs[slug] {
			notFound = append(notFound, slug)
		}
	}
	if len(notFound) > 0 {
		return fmt.Errorf("batch trash: files not found: %v", notFound)
	}

	// Validate all files
	var dirIDs []int64
	for _, f := range files {
		if f.IsSystem {
			return fmt.Errorf("batch trash: system file cannot be trashed: %s", f.Slug)
		}
		if err := ensureFileUnlocked(ctx, s.pg, userID, sessionID, f.ID); err != nil {
			return err
		}
		if f.IsDir {
			dirIDs = append(dirIDs, f.ID)
		}
	}

	// Check directories for non-empty
	if len(dirIDs) > 0 {
		var nonEmptyCount int64
		err = s.pg.QueryRow(ctx,
			`SELECT COUNT(*) FROM user_files
			 WHERE parent_id = ANY($1::bigint[]) AND is_trashed = FALSE`,
			dirIDs,
		).Scan(&nonEmptyCount)
		if err != nil {
			return fmt.Errorf("batch trash: count dir children: %w", err)
		}
		if nonEmptyCount > 0 {
			return model.ErrDirNotEmpty
		}
	}

	// Batch trash all files
	_, err = s.pg.Exec(ctx,
		`UPDATE user_files
		 SET is_trashed = TRUE, trashed_at = NOW(), updated_at = NOW()
		 WHERE slug = ANY($1::varchar[]) AND user_id = $2 AND is_trashed = FALSE`,
		slugs, userID,
	)
	if err != nil {
		return fmt.Errorf("batch trash: update: %w", err)
	}

	return nil
}

func (s *FilesService) RestoreFile(ctx context.Context, userID int64, fileSlug string) error {
	f, err := s.queries.GetFileBySlugForUser(ctx, sqlc.GetFileBySlugForUserParams{
		Slug:   fileSlug,
		UserID: userID,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return model.ErrNotFound
		}
		return fmt.Errorf("get file: %w", err)
	}
	if f.IsSystem {
		return model.ErrSystemFileLocked
	}
	if !f.IsTrashed {
		return nil
	}

	return s.queries.RestoreFile(ctx, f.ID)
}

func (s *FilesService) PermanentDelete(ctx context.Context, userID int64, fileSlug string) error {
	f, err := s.queries.GetFileBySlugForUser(ctx, sqlc.GetFileBySlugForUserParams{
		Slug:   fileSlug,
		UserID: userID,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return model.ErrNotFound
		}
		return fmt.Errorf("get file: %w", err)
	}
	if f.IsSystem {
		return model.ErrSystemFileLocked
	}

	// Decrement storage for non-directories
	if !f.IsDir && f.PhysicalFileID.Valid {
		_, err = s.queries.AtomicIncrementStorage(ctx, sqlc.AtomicIncrementStorageParams{
			UserID:      userID,
			StorageUsed: -f.FileSize,
		})
		if err != nil && !errors.Is(err, pgx.ErrNoRows) {
			return fmt.Errorf("decrement storage: %w", err)
		}
	}

	// Delete the user_file
	if err := s.queries.DeleteFile(ctx, f.ID); err != nil {
		return fmt.Errorf("delete file: %w", err)
	}

	// For files with physical backing, check if we should delete the physical file
	if !f.IsDir && f.PhysicalFileID.Valid {
		refCount, err := s.queries.CountReferencesByFileID(ctx, f.PhysicalFileID)
		if err != nil {
			return fmt.Errorf("count references: %w", err)
		}

		if refCount == 0 {
			pf, err := s.queries.GetPhysicalFileByID(ctx, f.PhysicalFileID.Int64)
			if err == nil {
				_ = s.store.Delete(pf.FileHash)
				_ = s.queries.DeletePhysicalFile(ctx, pf.ID)
			}
		}
	}

	return nil
}

func (s *FilesService) ForceDeleteDir(ctx context.Context, userID int64, sessionID, dirSlug string) error {
	dir, err := s.queries.GetFileBySlugForUser(ctx, sqlc.GetFileBySlugForUserParams{Slug: dirSlug, UserID: userID})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return model.ErrNotFound
		}
		return fmt.Errorf("get directory: %w", err)
	}
	if !dir.IsDir || dir.IsTrashed {
		return model.ErrNotFound
	}
	if dir.IsSystem {
		return model.ErrSystemFileLocked
	}
	if err := ensureFileUnlocked(ctx, s.pg, userID, sessionID, dir.ID); err != nil {
		return err
	}

	// Get all descendant file IDs and their physical file info via recursive CTE.
	rows, err := s.pg.Query(ctx, `
		WITH RECURSIVE subtree AS (
			SELECT id, is_dir, physical_file_id, file_size, 0 AS depth
			FROM user_files
			WHERE id = $1 AND user_id = $2
			UNION ALL
			SELECT f.id, f.is_dir, f.physical_file_id, f.file_size, s.depth + 1
			FROM user_files f
			JOIN subtree s ON s.id = f.parent_id
			WHERE f.user_id = $2 AND s.depth < 50
		)
		SELECT id, is_dir, physical_file_id, file_size FROM subtree ORDER BY depth DESC
	`, dir.ID, userID)
	if err != nil {
		return fmt.Errorf("query subtree: %w", err)
	}
	defer rows.Close()

	type subtreeFile struct {
		ID             int64
		IsDir          bool
		PhysicalFileID pgtype.Int8
		FileSize       int64
	}
	var allFiles []subtreeFile
	var fileIDs []int64
	var reclaimSize int64
	var physicalIDs []int64

	for rows.Next() {
		var f subtreeFile
		if err := rows.Scan(&f.ID, &f.IsDir, &f.PhysicalFileID, &f.FileSize); err != nil {
			return fmt.Errorf("scan subtree file: %w", err)
		}
		allFiles = append(allFiles, f)
		fileIDs = append(fileIDs, f.ID)
		if !f.IsDir && f.PhysicalFileID.Valid {
			reclaimSize += f.FileSize
			physicalIDs = append(physicalIDs, f.PhysicalFileID.Int64)
		}
	}
	if err := rows.Err(); err != nil {
		return err
	}
	if len(fileIDs) == 0 {
		return model.ErrNotFound
	}

	// Delete all subtree files in a single query.
	_, err = s.pg.Exec(ctx, `DELETE FROM user_files WHERE id = ANY($1::bigint[]) AND user_id = $2`, fileIDs, userID)
	if err != nil {
		return fmt.Errorf("delete subtree: %w", err)
	}

	// Decrement storage.
	if reclaimSize > 0 {
		if _, err := s.queries.AtomicIncrementStorage(ctx, sqlc.AtomicIncrementStorageParams{
			UserID:      userID,
			StorageUsed: -reclaimSize,
		}); err != nil {
			slog.Warn("force delete: decrement storage", "userID", userID, "error", err)
		}
	}

	// Clean up unreferenced physical files.
	for _, pfID := range physicalIDs {
		refCount, err := s.queries.CountReferencesByFileID(ctx, pgtype.Int8{Int64: pfID, Valid: true})
		if err != nil {
			continue
		}
		if refCount == 0 {
			pf, err := s.queries.GetPhysicalFileByID(ctx, pfID)
			if err == nil {
				_ = s.store.Delete(pf.FileHash)
				_ = s.queries.DeletePhysicalFile(ctx, pf.ID)
			}
		}
	}

	return nil
}

func (s *FilesService) EmptyTrash(ctx context.Context, userID int64) (int, error) {
	// Get all trashed files
	rows, err := s.pg.Query(ctx,
		"SELECT id, is_dir, physical_file_id, file_size FROM user_files WHERE user_id = $1 AND is_trashed = TRUE AND is_system = FALSE",
		userID,
	)
	if err != nil {
		return 0, fmt.Errorf("list trashed files: %w", err)
	}
	defer rows.Close()

	type trashedFile struct {
		ID             int64
		IsDir          bool
		PhysicalFileID pgtype.Int8
		FileSize       int64
	}

	var files []trashedFile
	for rows.Next() {
		var f trashedFile
		if err := rows.Scan(&f.ID, &f.IsDir, &f.PhysicalFileID, &f.FileSize); err != nil {
			return 0, fmt.Errorf("scan trashed file: %w", err)
		}
		files = append(files, f)
	}
	if err := rows.Err(); err != nil {
		return 0, err
	}

	if len(files) == 0 {
		return 0, nil
	}

	// Calculate total size to reclaim
	var reclaimSize int64
	var physicalIDs []int64
	for _, f := range files {
		if !f.IsDir && f.PhysicalFileID.Valid {
			reclaimSize += f.FileSize
			physicalIDs = append(physicalIDs, f.PhysicalFileID.Int64)
		}
	}

	// Delete all trashed files
	_, err = s.pg.Exec(ctx, "DELETE FROM user_files WHERE user_id = $1 AND is_trashed = TRUE AND is_system = FALSE", userID)
	if err != nil {
		return 0, fmt.Errorf("delete trashed files: %w", err)
	}

	// Decrement storage
	if reclaimSize > 0 {
		if _, err := s.queries.AtomicIncrementStorage(ctx, sqlc.AtomicIncrementStorageParams{
			UserID:      userID,
			StorageUsed: -reclaimSize,
		}); err != nil {
			// Log error but don't fail - user files are already deleted
			// Storage quota may be slightly inaccurate until next recalculation
		}
	}

	// Clean up unreferenced physical files
	for _, pfID := range physicalIDs {
		refCount, err := s.queries.CountReferencesByFileID(ctx, pgtype.Int8{Int64: pfID, Valid: true})
		if err != nil {
			continue
		}
		if refCount == 0 {
			pf, err := s.queries.GetPhysicalFileByID(ctx, pfID)
			if err == nil {
				_ = s.store.Delete(pf.FileHash)
				_ = s.queries.DeletePhysicalFile(ctx, pf.ID)
			}
		}
	}

	return len(files), nil
}

func (s *FilesService) RestoreAll(ctx context.Context, userID int64) (int, error) {
	result, err := s.pg.Exec(ctx,
		"UPDATE user_files SET is_trashed = FALSE, trashed_at = NULL, updated_at = NOW() WHERE user_id = $1 AND is_trashed = TRUE AND is_system = FALSE",
		userID,
	)
	if err != nil {
		return 0, fmt.Errorf("restore all files: %w", err)
	}

	return int(result.RowsAffected()), nil
}

func (s *FilesService) RenameFile(ctx context.Context, userID int64, sessionID, fileSlug, newName string) error {
	if newName == "" || len(newName) > 100 {
		return model.ErrInvalidInput
	}

	f, err := s.queries.GetFileBySlugForUser(ctx, sqlc.GetFileBySlugForUserParams{
		Slug:   fileSlug,
		UserID: userID,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return model.ErrNotFound
		}
		return fmt.Errorf("get file: %w", err)
	}
	if f.IsSystem {
		return model.ErrSystemFileLocked
	}
	if err := ensureFileUnlocked(ctx, s.pg, userID, sessionID, f.ID); err != nil {
		return err
	}

	conflict, err := s.queries.CheckNameConflict(ctx, sqlc.CheckNameConflictParams{
		UserID:   userID,
		ParentID: f.ParentID,
		FileName: newName,
		IsSystem: false,
	})
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return fmt.Errorf("check conflict: %w", err)
	}
	if conflict.ID != 0 && conflict.ID != f.ID {
		return model.ErrNameConflict
	}

	return s.queries.RenameFile(ctx, sqlc.RenameFileParams{
		ID:       f.ID,
		FileName: newName,
	})
}

func (s *FilesService) MoveFile(ctx context.Context, userID int64, sessionID, fileSlug, targetParentSlug string) error {
	f, err := s.queries.GetFileBySlugForUser(ctx, sqlc.GetFileBySlugForUserParams{
		Slug:   fileSlug,
		UserID: userID,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return model.ErrNotFound
		}
		return fmt.Errorf("get file: %w", err)
	}
	if f.IsSystem {
		return model.ErrSystemFileLocked
	}
	if err := ensureFileUnlocked(ctx, s.pg, userID, sessionID, f.ID); err != nil {
		return err
	}

	var targetParentID pgtype.Int8
	if targetParentSlug != "" {
		target, err := s.queries.GetFileBySlugForUser(ctx, sqlc.GetFileBySlugForUserParams{
			Slug:   targetParentSlug,
			UserID: userID,
		})
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return model.ErrNotFound
			}
			return fmt.Errorf("get target parent: %w", err)
		}
		if !target.IsDir {
			return model.ErrInvalidInput
		}
		if err := ensureFileUnlocked(ctx, s.pg, userID, sessionID, target.ID); err != nil {
			return err
		}
		if f.IsDir {
			if target.ID == f.ID {
				return model.ErrInvalidInput
			}
			isDescendant, err := s.isDescendantDir(ctx, userID, f.ID, target.ID)
			if err != nil {
				return err
			}
			if isDescendant {
				return model.ErrInvalidInput
			}
		}
		targetParentID = pgtype.Int8{Int64: target.ID, Valid: true}
	}

	conflict, err := s.queries.CheckNameConflict(ctx, sqlc.CheckNameConflictParams{
		UserID:   userID,
		ParentID: targetParentID,
		FileName: f.FileName,
		IsSystem: f.IsSystem,
	})
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return fmt.Errorf("check conflict: %w", err)
	}
	if conflict.ID != 0 && conflict.ID != f.ID {
		return model.ErrNameConflict
	}

	return s.queries.MoveFile(ctx, sqlc.MoveFileParams{
		ID:         f.ID,
		ParentID:   targetParentID,
		ParentSlug: pgtype.Text{String: targetParentSlug, Valid: targetParentSlug != ""},
	})
}

func (s *FilesService) isDescendantDir(ctx context.Context, userID, sourceID, targetID int64) (bool, error) {
	for currentID := targetID; ; {
		if currentID == sourceID {
			return true, nil
		}

		current, err := s.queries.GetFileByID(ctx, currentID)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return false, model.ErrNotFound
			}
			return false, fmt.Errorf("get parent chain file: %w", err)
		}
		if current.UserID != userID {
			return false, model.ErrNotFound
		}
		if !current.ParentID.Valid {
			return false, nil
		}
		currentID = current.ParentID.Int64
	}
}

func (s *FilesService) SetStarred(ctx context.Context, userID int64, sessionID, fileSlug string, starred bool) error {
	f, err := s.queries.GetFileBySlugForUser(ctx, sqlc.GetFileBySlugForUserParams{
		Slug:   fileSlug,
		UserID: userID,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return model.ErrNotFound
		}
		return fmt.Errorf("get file: %w", err)
	}
	if f.IsSystem {
		return model.ErrSystemFileLocked
	}
	if err := ensureFileUnlocked(ctx, s.pg, userID, sessionID, f.ID); err != nil {
		return err
	}

	return s.queries.SetStarred(ctx, sqlc.SetStarredParams{
		ID:        f.ID,
		IsStarred: starred,
	})
}

type DownloadFileResult struct {
	File     *os.File
	Name     string
	MimeType string
	FileHash string
}

func (s *FilesService) DownloadFile(ctx context.Context, userID int64, sessionID, fileSlug string) (*DownloadFileResult, error) {
	f, err := s.queries.GetFileBySlugForUser(ctx, sqlc.GetFileBySlugForUserParams{
		Slug:   fileSlug,
		UserID: userID,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, model.ErrNotFound
		}
		return nil, fmt.Errorf("get file: %w", err)
	}
	if f.IsDir || f.IsTrashed || !f.PhysicalFileID.Valid {
		return nil, model.ErrNotFound
	}
	if err := ensureFileUnlocked(ctx, s.pg, userID, sessionID, f.ID); err != nil {
		return nil, err
	}

	pf, err := s.queries.GetPhysicalFileByID(ctx, f.PhysicalFileID.Int64)
	if err != nil {
		return nil, fmt.Errorf("get physical file: %w", err)
	}

	file, err := s.store.Open(pf.FileHash)
	if err != nil {
		return nil, fmt.Errorf("open file: %w", err)
	}

	mimeType := "application/octet-stream"
	if f.MimeType.Valid && f.MimeType.String != "" {
		mimeType = f.MimeType.String
	}

	return &DownloadFileResult{
		File:     file,
		Name:     safeFilename(f.FileName),
		MimeType: mimeType,
		FileHash: pf.FileHash,
	}, nil
}

func fileRowsToItems(files []db.FileRow) []FileItem {
	items := make([]FileItem, 0, len(files))
	for _, f := range files {
		item := FileItem{
			Slug:         f.Slug,
			FileName:     f.FileName,
			IsDir:        f.IsDir,
			HasPassword:  f.LockPasswordHash.Valid && f.LockPasswordHash.String != "",
			FileSize:     f.FileSize,
			FileCategory: f.FileCategory,
			IsStarred:    f.IsStarred,
			IsSystem:     f.IsSystem,
			CreatedAt:    f.CreatedAt.Time.Format("2006-01-02T15:04:05Z"),
			UpdatedAt:    f.UpdatedAt.Time.Format("2006-01-02T15:04:05Z"),
		}
		if f.MimeType.Valid {
			item.MimeType = &f.MimeType.String
		}
		if f.SystemKind.Valid {
			item.SystemKind = &f.SystemKind.String
		}
		if f.ParentSlug.Valid {
			item.ParentSlug = &f.ParentSlug.String
		}
		if f.ParentName.Valid {
			item.ParentName = &f.ParentName.String
		}
		if f.HashAlgo.Valid {
			item.HashAlgo = &f.HashAlgo.String
		}
		if f.FileHash.Valid {
			item.FileHash = &f.FileHash.String
		}
		items = append(items, item)
	}
	return items
}

func (s *FilesService) resolveParentID(ctx context.Context, userID int64, parentSlug string) (pgtype.Int8, error) {
	if parentSlug == "" {
		return pgtype.Int8{}, nil
	}

	parent, err := s.queries.GetFileBySlugForUser(ctx, sqlc.GetFileBySlugForUserParams{
		Slug:   parentSlug,
		UserID: userID,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return pgtype.Int8{}, model.ErrNotFound
		}
		return pgtype.Int8{}, fmt.Errorf("get parent: %w", err)
	}
	if parent.IsTrashed || !parent.IsDir {
		return pgtype.Int8{}, model.ErrNotFound
	}
	return pgtype.Int8{Int64: parent.ID, Valid: true}, nil
}

func (s *FilesService) ensureResolvedParentUnlocked(ctx context.Context, userID int64, sessionID string, parentID pgtype.Int8) error {
	if !parentID.Valid {
		return nil
	}
	return ensureFileUnlocked(ctx, s.pg, userID, sessionID, parentID.Int64)
}

func fileToItem(f sqlc.UserFile) FileItem {
	item := FileItem{
		Slug:         f.Slug,
		FileName:     f.FileName,
		IsDir:        f.IsDir,
		HasPassword:  f.LockPasswordHash.Valid && f.LockPasswordHash.String != "",
		FileSize:     f.FileSize,
		FileCategory: f.FileCategory,
		IsStarred:    f.IsStarred,
		IsSystem:     f.IsSystem,
		CreatedAt:    f.CreatedAt.Time.Format("2006-01-02T15:04:05Z"),
		UpdatedAt:    f.UpdatedAt.Time.Format("2006-01-02T15:04:05Z"),
	}
	if f.MimeType.Valid {
		item.MimeType = &f.MimeType.String
	}
	if f.ParentSlug.Valid {
		item.ParentSlug = &f.ParentSlug.String
	}
	if f.SystemKind.Valid {
		item.SystemKind = &f.SystemKind.String
	}
	return item
}

func safeFilename(name string) string {
	safe := strings.Map(func(r rune) rune {
		if r > 127 || r == '/' || r == '\\' || r == 0 {
			return '_'
		}
		return r
	}, name)
	if safe == "" || safe == "." {
		safe = "download"
	}
	return safe
}

func contentDisposition(filename string) string {
	return fmt.Sprintf(`attachment; filename="%s"; filename*=UTF-8''%s`, safeFilename(filename), filename)
}

type FolderSummary struct {
	FileCount   int64 `json:"fileCount"`
	FolderCount int64 `json:"folderCount"`
	TotalSize   int64 `json:"totalSize"`
}

func (s *FilesService) GetFolderSummary(ctx context.Context, userID int64, slug string) (*FolderSummary, error) {
	f, err := s.queries.GetFileBySlugForUser(ctx, sqlc.GetFileBySlugForUserParams{
		Slug:   slug,
		UserID: userID,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, model.ErrNotFound
		}
		return nil, fmt.Errorf("get folder: %w", err)
	}
	if !f.IsDir {
		return nil, model.ErrInvalidInput
	}

	var summary FolderSummary
	err = s.pg.QueryRow(ctx, `
		SELECT
			COUNT(*) FILTER (WHERE is_dir = false) AS file_count,
			COUNT(*) FILTER (WHERE is_dir = true) AS folder_count,
			COALESCE(SUM(file_size) FILTER (WHERE is_dir = false), 0) AS total_size
		FROM user_files
		WHERE parent_id = $1 AND is_trashed = false
	`, f.ID).Scan(&summary.FileCount, &summary.FolderCount, &summary.TotalSize)
	if err != nil {
		return nil, fmt.Errorf("query folder summary: %w", err)
	}

	return &summary, nil
}

func detectMime(path string) string {
	f, err := os.Open(path)
	if err != nil {
		return "application/octet-stream"
	}
	defer f.Close()

	buf := make([]byte, 512)
	n, err := f.Read(buf)
	if err != nil {
		return "application/octet-stream"
	}
	return http.DetectContentType(buf[:n])
}
