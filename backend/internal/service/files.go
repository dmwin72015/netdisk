package service

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

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

type FileItem struct {
	Slug         string  `json:"slug"`
	FileName     string  `json:"fileName"`
	IsDir        bool    `json:"isDir"`
	FileSize     int64   `json:"fileSize"`
	MimeType     *string `json:"mimeType"`
	FileCategory string  `json:"fileCategory"`
	IsStarred    bool    `json:"isStarred"`
	ParentSlug   *string `json:"parentSlug,omitempty"`
	ParentName   *string `json:"parentName,omitempty"`
	CreatedAt    string  `json:"createdAt"`
	UpdatedAt    string  `json:"updatedAt"`
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

// ResolveParent looks up a parent directory by slug, verifying it exists and is not trashed.
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
	return parent, nil
}

// ListUserFiles is the unified file listing method powered by Squirrel.
func (s *FilesService) ListUserFiles(ctx context.Context, params db.ListFilesParams) ([]FileItem, int, error) {
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

	return fileRowsToItems(fileRows), total, nil
}

func (s *FilesService) ListRecentFiles(ctx context.Context, userID int64, limit int) ([]FileItem, int, error) {
	params := db.ListFilesParams{
		UserID:         userID,
		IncludeDirs:    false,
		IgnoreParentID: true, // show files from all directories
		SortBy:         "created_at",
		SortDir:        "DESC",
		Page:           1,
		PageSize:       limit,
	}

	return s.ListUserFiles(ctx, params)
}

type BreadcrumbItem struct {
	Slug     string `json:"slug"`
	FileName string `json:"fileName"`
}

func (s *FilesService) GetBreadcrumb(ctx context.Context, userID int64, slug string) ([]BreadcrumbItem, error) {
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

func (s *FilesService) Mkdir(ctx context.Context, userID int64, dirName, parentSlug string) (*FileItem, error) {
	if dirName == "" || len(dirName) > 100 {
		return nil, model.ErrInvalidInput
	}

	var parentID pgtype.Int8
	if parentSlug != "" {
		parent, err := s.queries.GetFileBySlugForUser(ctx, sqlc.GetFileBySlugForUserParams{
			Slug:   parentSlug,
			UserID: userID,
		})
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return nil, model.ErrNotFound
			}
			return nil, fmt.Errorf("get parent: %w", err)
		}
		parentID = pgtype.Int8{Int64: parent.ID, Valid: true}
	}

	conflict, err := s.queries.CheckNameConflict(ctx, sqlc.CheckNameConflictParams{
		UserID:   userID,
		ParentID: parentID,
		FileName: dirName,
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

func (s *FilesService) CheckConflict(ctx context.Context, userID int64, fileName, preHash, parentSlug string, fileSize int64) (*ConflictResponse, error) {
	var parentID pgtype.Int8
	if parentSlug != "" {
		parent, err := s.queries.GetFileBySlugForUser(ctx, sqlc.GetFileBySlugForUserParams{
			Slug:   parentSlug,
			UserID: userID,
		})
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return nil, model.ErrNotFound
			}
			return nil, fmt.Errorf("get parent: %w", err)
		}
		parentID = pgtype.Int8{Int64: parent.ID, Valid: true}
	}

	conflict, err := s.queries.CheckNameConflict(ctx, sqlc.CheckNameConflictParams{
		UserID:   userID,
		ParentID: parentID,
		FileName: fileName,
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

func (s *FilesService) CheckDuplicate(ctx context.Context, userID int64, fileHash, parentSlug string) (*ConflictResponse, error) {
	if parentSlug != "" {
		_, err := s.queries.GetFileBySlugForUser(ctx, sqlc.GetFileBySlugForUserParams{
			Slug:   parentSlug,
			UserID: userID,
		})
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return nil, model.ErrNotFound
			}
			return nil, fmt.Errorf("get parent: %w", err)
		}
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

func (s *FilesService) ImportFile(ctx context.Context, userID int64, physicalFileSlug, fileName, parentSlug string) (*ImportResponse, error) {
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

	var parentID pgtype.Int8
	if parentSlug != "" {
		parent, err := s.queries.GetFileBySlugForUser(ctx, sqlc.GetFileBySlugForUserParams{
			Slug:   parentSlug,
			UserID: userID,
		})
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return nil, model.ErrNotFound
			}
			return nil, fmt.Errorf("get parent: %w", err)
		}
		parentID = pgtype.Int8{Int64: parent.ID, Valid: true}
	}

	conflict, err := s.queries.CheckNameConflict(ctx, sqlc.CheckNameConflictParams{
		UserID:   userID,
		ParentID: parentID,
		FileName: fileName,
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
	})
	if err != nil {
		return nil, fmt.Errorf("create file: %w", err)
	}

	return &ImportResponse{
		FileSlug: f.Slug,
		FileName: f.FileName,
	}, nil
}

func (s *FilesService) TrashFile(ctx context.Context, userID int64, fileSlug string) error {
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

func (s *FilesService) EmptyTrash(ctx context.Context, userID int64) (int, error) {
	// Get all trashed files
	rows, err := s.pg.Query(ctx,
		"SELECT id, is_dir, physical_file_id, file_size FROM user_files WHERE user_id = $1 AND is_trashed = TRUE",
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
	_, err = s.pg.Exec(ctx, "DELETE FROM user_files WHERE user_id = $1 AND is_trashed = TRUE", userID)
	if err != nil {
		return 0, fmt.Errorf("delete trashed files: %w", err)
	}

	// Decrement storage
	if reclaimSize > 0 {
		_, _ = s.queries.AtomicIncrementStorage(ctx, sqlc.AtomicIncrementStorageParams{
			UserID:      userID,
			StorageUsed: -reclaimSize,
		})
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
		"UPDATE user_files SET is_trashed = FALSE, trashed_at = NULL, updated_at = NOW() WHERE user_id = $1 AND is_trashed = TRUE",
		userID,
	)
	if err != nil {
		return 0, fmt.Errorf("restore all files: %w", err)
	}

	return int(result.RowsAffected()), nil
}

func (s *FilesService) RenameFile(ctx context.Context, userID int64, fileSlug, newName string) error {
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

	conflict, err := s.queries.CheckNameConflict(ctx, sqlc.CheckNameConflictParams{
		UserID:   userID,
		ParentID: f.ParentID,
		FileName: newName,
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

func (s *FilesService) MoveFile(ctx context.Context, userID int64, fileSlug, targetParentSlug string) error {
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
		targetParentID = pgtype.Int8{Int64: target.ID, Valid: true}
	}

	conflict, err := s.queries.CheckNameConflict(ctx, sqlc.CheckNameConflictParams{
		UserID:   userID,
		ParentID: targetParentID,
		FileName: f.FileName,
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

func (s *FilesService) SetStarred(ctx context.Context, userID int64, fileSlug string, starred bool) error {
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

	return s.queries.SetStarred(ctx, sqlc.SetStarredParams{
		ID:        f.ID,
		IsStarred: starred,
	})
}


func (s *FilesService) DownloadFile(ctx context.Context, userID int64, fileSlug string) (io.ReadSeeker, string, string, error) {
	f, err := s.queries.GetFileBySlugForUser(ctx, sqlc.GetFileBySlugForUserParams{
		Slug:   fileSlug,
		UserID: userID,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, "", "", model.ErrNotFound
		}
		return nil, "", "", fmt.Errorf("get file: %w", err)
	}
	if f.IsDir || f.IsTrashed || !f.PhysicalFileID.Valid {
		return nil, "", "", model.ErrNotFound
	}

	pf, err := s.queries.GetPhysicalFileByID(ctx, f.PhysicalFileID.Int64)
	if err != nil {
		return nil, "", "", fmt.Errorf("get physical file: %w", err)
	}

	file, err := s.store.Open(pf.FileHash)
	if err != nil {
		return nil, "", "", fmt.Errorf("open file: %w", err)
	}

	mimeType := "application/octet-stream"
	if f.MimeType.Valid && f.MimeType.String != "" {
		mimeType = f.MimeType.String
	}

	safeName := safeFilename(f.FileName)
	return file, safeName, mimeType, nil
}

func fileRowsToItems(files []db.FileRow) []FileItem {
	items := make([]FileItem, 0, len(files))
	for _, f := range files {
		item := FileItem{
			Slug:         f.Slug,
			FileName:     f.FileName,
			IsDir:        f.IsDir,
			FileSize:     f.FileSize,
			FileCategory: f.FileCategory,
			IsStarred:    f.IsStarred,
			CreatedAt:    f.CreatedAt.Time.Format("2006-01-02T15:04:05Z"),
			UpdatedAt:    f.UpdatedAt.Time.Format("2006-01-02T15:04:05Z"),
		}
		if f.MimeType.Valid {
			item.MimeType = &f.MimeType.String
		}
		if f.ParentSlug.Valid {
			item.ParentSlug = &f.ParentSlug.String
		}
		if f.ParentName.Valid {
			item.ParentName = &f.ParentName.String
		}
		items = append(items, item)
	}
	return items
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
