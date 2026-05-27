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
	"github.com/netdisk/server/internal/db/sqlc"
	"github.com/netdisk/server/internal/model"
	"github.com/netdisk/server/internal/storage"
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
	Slug      string  `json:"slug"`
	FileName  string  `json:"file_name"`
	IsDir     bool    `json:"is_dir"`
	FileSize  int64   `json:"file_size"`
	MimeType  *string `json:"mime_type"`
	IsStarred bool    `json:"is_starred"`
	CreatedAt string  `json:"created_at"`
	UpdatedAt string  `json:"updated_at"`
}

type ConflictResponse struct {
	Status   string     `json:"status"`
	Message  string     `json:"message,omitempty"`
	Existing *FileItem  `json:"existing,omitempty"`
}

type ImportResponse struct {
	FileSlug string `json:"file_slug"`
	FileName string `json:"file_name"`
}

func (s *FilesService) ListFiles(ctx context.Context, userID int64, parentSlug string, page, pageSize int) ([]FileItem, int, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 50
	}

	var parentID pgtype.Int8
	if parentSlug != "" {
		parent, err := s.queries.GetFileBySlugForUser(ctx, sqlc.GetFileBySlugForUserParams{
			Slug:   parentSlug,
			UserID: userID,
		})
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return nil, 0, model.ErrNotFound
			}
			return nil, 0, fmt.Errorf("get parent: %w", err)
		}
		if parent.IsTrashed {
			return nil, 0, model.ErrNotFound
		}
		parentID = pgtype.Int8{Int64: parent.ID, Valid: true}
	}

	count, err := s.queries.CountFilesByParent(ctx, sqlc.CountFilesByParentParams{
		UserID:   userID,
		ParentID: parentID,
	})
	if err != nil {
		return nil, 0, fmt.Errorf("count files: %w", err)
	}

	files, err := s.queries.ListFilesByParent(ctx, sqlc.ListFilesByParentParams{
		UserID:     userID,
		ParentID:   parentID,
		PageSize:   int32(pageSize),
		PageOffset: int32((page - 1) * pageSize),
	})
	if err != nil {
		return nil, 0, fmt.Errorf("list files: %w", err)
	}

	items := make([]FileItem, 0, len(files))
	for _, f := range files {
		item := FileItem{
			Slug:      f.Slug,
			FileName:  f.FileName,
			IsDir:     f.IsDir,
			FileSize:  f.FileSize,
			IsStarred: f.IsStarred,
			CreatedAt: f.CreatedAt.Time.Format("2006-01-02T15:04:05Z"),
			UpdatedAt: f.UpdatedAt.Time.Format("2006-01-02T15:04:05Z"),
		}
		if f.MimeType.Valid {
			item.MimeType = &f.MimeType.String
		}
		items = append(items, item)
	}

	return items, int(count), nil
}

func (s *FilesService) ListFilesByMime(ctx context.Context, userID int64, mimePrefix string, page, pageSize int) ([]FileItem, int, error) {
	if mimePrefix == "" {
		return nil, 0, model.ErrInvalidInput
	}
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 50
	}

	mimePattern := pgtype.Text{String: mimePrefix + "%", Valid: true}

	count, err := s.queries.CountFilesByMimePrefix(ctx, sqlc.CountFilesByMimePrefixParams{
		UserID:     userID,
		MimePrefix: mimePattern,
	})
	if err != nil {
		return nil, 0, fmt.Errorf("count files by mime: %w", err)
	}

	files, err := s.queries.ListFilesByMimePrefix(ctx, sqlc.ListFilesByMimePrefixParams{
		UserID:     userID,
		MimePrefix: mimePattern,
		PageSize:   int32(pageSize),
		PageOffset: int32((page - 1) * pageSize),
	})
	if err != nil {
		return nil, 0, fmt.Errorf("list files by mime: %w", err)
	}

	return fileItemsFromRows(files), int(count), nil
}

type BreadcrumbItem struct {
	Slug     string `json:"slug"`
	FileName string `json:"file_name"`
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
		Slug:     slug,
		UserID:   userID,
		ParentID: parentID,
		FileName: dirName,
		IsDir:    true,
		FileSize: 0,
	})
	if err != nil {
		if isUniqueViolation(err) {
			return nil, model.ErrNameConflict
		}
		return nil, fmt.Errorf("create dir: %w", err)
	}

	return &FileItem{
		Slug:      f.Slug,
		FileName:  f.FileName,
		IsDir:     true,
		FileSize:  0,
		CreatedAt: f.CreatedAt.Time.Format("2006-01-02T15:04:05Z"),
		UpdatedAt: f.UpdatedAt.Time.Format("2006-01-02T15:04:05Z"),
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

	f, err := s.queries.CreateFile(ctx, sqlc.CreateFileParams{
		Slug:           slug,
		UserID:         userID,
		PhysicalFileID: pgtype.Int8{Int64: pf.ID, Valid: true},
		ParentID:       parentID,
		FileName:       fileName,
		IsDir:          false,
		FileSize:       pf.FileSize,
		MimeType:       mimeType,
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
		ID:       f.ID,
		ParentID: targetParentID,
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

func (s *FilesService) ListTrashed(ctx context.Context, userID int64, page, pageSize int) ([]FileItem, int, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 50
	}

	total, err := s.queries.CountTrashedFiles(ctx, userID)
	if err != nil {
		return nil, 0, fmt.Errorf("count trashed: %w", err)
	}

	files, err := s.queries.ListTrashedFiles(ctx, sqlc.ListTrashedFilesParams{
		UserID: userID,
		Limit:  int32(pageSize),
		Offset: int32((page - 1) * pageSize),
	})
	if err != nil {
		return nil, 0, fmt.Errorf("list trashed: %w", err)
	}

	return fileItemsFromRows(files), int(total), nil
}

func (s *FilesService) ListStarred(ctx context.Context, userID int64, page, pageSize int) ([]FileItem, int, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 50
	}

	total, err := s.queries.CountStarredFiles(ctx, userID)
	if err != nil {
		return nil, 0, fmt.Errorf("count starred: %w", err)
	}

	files, err := s.queries.ListStarredFiles(ctx, sqlc.ListStarredFilesParams{
		UserID: userID,
		Limit:  int32(pageSize),
		Offset: int32((page - 1) * pageSize),
	})
	if err != nil {
		return nil, 0, fmt.Errorf("list starred: %w", err)
	}

	return fileItemsFromRows(files), int(total), nil
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

func fileItemsFromRows(files []sqlc.UserFile) []FileItem {
	items := make([]FileItem, 0, len(files))
	for _, f := range files {
		item := FileItem{
			Slug:      f.Slug,
			FileName:  f.FileName,
			IsDir:     f.IsDir,
			FileSize:  f.FileSize,
			IsStarred: f.IsStarred,
			CreatedAt: f.CreatedAt.Time.Format("2006-01-02T15:04:05Z"),
			UpdatedAt: f.UpdatedAt.Time.Format("2006-01-02T15:04:05Z"),
		}
		if f.MimeType.Valid {
			item.MimeType = &f.MimeType.String
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
