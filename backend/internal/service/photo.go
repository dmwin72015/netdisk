package service

import (
	"context"
	"errors"
	"fmt"
	"math"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	gonanoid "github.com/matoous/go-nanoid/v2"
	"github.com/rs/zerolog"

	"github.com/netdisk/server/internal/config"
	"github.com/netdisk/server/internal/db/sqlc"
	"github.com/netdisk/server/internal/model"
	"github.com/netdisk/server/internal/storage"
)

type PhotoService struct {
	queries *sqlc.Queries
	pg      *pgxpool.Pool
	cfg     *config.Config
	store   *storage.Local
	logger  zerolog.Logger
}

func NewPhotoService(queries *sqlc.Queries, pg *pgxpool.Pool, cfg *config.Config, store *storage.Local, logger zerolog.Logger) *PhotoService {
	return &PhotoService{
		queries: queries,
		pg:      pg,
		cfg:     cfg,
		store:   store,
		logger:  logger,
	}
}

type PhotoItem struct {
	Slug        string  `json:"slug"`
	FileName    string  `json:"fileName"`
	FileSize    int64   `json:"fileSize"`
	MimeType    *string `json:"mimeType"`
	IsStarred   bool    `json:"isStarred"`
	CreatedAt   string  `json:"createdAt"`
	UpdatedAt   string  `json:"updatedAt"`
	ThumbnailURL string `json:"thumbnailUrl"`
	FileHash    *string `json:"fileHash"`
}

type PhotoListResponse struct {
	Items []PhotoItem `json:"items"`
	Total int         `json:"total"`
	Page  int         `json:"page"`
}

func (s *PhotoService) ListPhotos(ctx context.Context, userID int64, page, pageSize int) (*PhotoListResponse, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 50
	}

	total, err := s.queries.CountImageFiles(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("count images: %w", err)
	}

	rows, err := s.queries.ListImageFiles(ctx, sqlc.ListImageFilesParams{
		UserID: userID,
		Limit:  int32(pageSize),
		Offset: int32((page - 1) * pageSize),
	})
	if err != nil {
		return nil, fmt.Errorf("list images: %w", err)
	}

	items := make([]PhotoItem, 0, len(rows))
	for _, r := range rows {
		item := PhotoItem{
			Slug:      r.Slug,
			FileName:  r.FileName,
			FileSize:  r.FileSize,
			IsStarred: r.IsStarred,
			CreatedAt: r.CreatedAt.Time.Format("2006-01-02T15:04:05Z"),
			UpdatedAt: r.UpdatedAt.Time.Format("2006-01-02T15:04:05Z"),
		}
		if r.MimeType.Valid {
			m := r.MimeType.String
			item.MimeType = &m
		}
		if r.FileHash.Valid {
			h := r.FileHash.String
			item.FileHash = &h
			item.ThumbnailURL = fmt.Sprintf("/api/v1/photos/%s/thumbnail", r.Slug)
		}
		items = append(items, item)
	}

	return &PhotoListResponse{
		Items: items,
		Total: int(total),
		Page:  page,
	}, nil
}

func (s *PhotoService) GetThumbnailPath(ctx context.Context, userID int64, fileSlug string) (string, error) {
	uf, err := s.queries.GetFileBySlugForUser(ctx, sqlc.GetFileBySlugForUserParams{
		Slug:   fileSlug,
		UserID: userID,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return "", model.ErrNotFound
		}
		return "", fmt.Errorf("get file: %w", err)
	}

	if !uf.PhysicalFileID.Valid {
		return "", model.ErrNotFound
	}

	pf, err := s.queries.GetPhysicalFileByID(ctx, uf.PhysicalFileID.Int64)
	if err != nil {
		return "", fmt.Errorf("get physical file: %w", err)
	}

	thumbPath, err := storage.ServeThumbnail(s.cfg.Storage.Root, s.cfg.Storage.FilesDir, pf.FileHash)
	if err != nil {
		return "", fmt.Errorf("serve thumbnail: %w", err)
	}

	return thumbPath, nil
}

func (s *PhotoService) GetPhotoDetail(ctx context.Context, userID int64, fileSlug string) (*PhotoItem, error) {
	uf, err := s.queries.GetFileBySlugForUser(ctx, sqlc.GetFileBySlugForUserParams{
		Slug:   fileSlug,
		UserID: userID,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, model.ErrNotFound
		}
		return nil, fmt.Errorf("get file: %w", err)
	}

	if uf.IsDir || uf.IsTrashed || uf.FileCategory != "image" {
		return nil, model.ErrNotFound
	}

	item := PhotoItem{
		Slug:      uf.Slug,
		FileName:  uf.FileName,
		FileSize:  uf.FileSize,
		IsStarred: uf.IsStarred,
		CreatedAt: uf.CreatedAt.Time.Format("2006-01-02T15:04:05Z"),
		UpdatedAt: uf.UpdatedAt.Time.Format("2006-01-02T15:04:05Z"),
	}
	if uf.MimeType.Valid {
		m := uf.MimeType.String
		item.MimeType = &m
		item.ThumbnailURL = fmt.Sprintf("/api/v1/photos/%s/thumbnail", uf.Slug)
	}

	if uf.PhysicalFileID.Valid {
		pf, err := s.queries.GetPhysicalFileByID(ctx, uf.PhysicalFileID.Int64)
		if err == nil {
			h := pf.FileHash
			item.FileHash = &h
		}
	}

	return &item, nil
}

type AlbumResponse struct {
	Slug        string  `json:"slug"`
	Title       string  `json:"title"`
	Description string  `json:"description"`
	CoverURL    *string `json:"coverUrl"`
	ItemCount   int     `json:"itemCount"`
	CreatedAt   string  `json:"createdAt"`
	UpdatedAt   string  `json:"updatedAt"`
}

type AlbumListResponse struct {
	Items []AlbumResponse `json:"items"`
	Total int             `json:"total"`
}

type AlbumCreateRequest struct {
	Title       string `json:"title"`
	Description string `json:"description"`
}

type AlbumUpdateRequest struct {
	Title         string  `json:"title"`
	Description   string  `json:"description"`
	CoverFileSlug *string `json:"coverFileSlug"`
}

func (s *PhotoService) CreateAlbum(ctx context.Context, userID int64, req AlbumCreateRequest) (*AlbumResponse, error) {
	if req.Title == "" {
		return nil, model.ErrInvalidInput
	}

	slug, err := gonanoid.New(21)
	if err != nil {
		return nil, fmt.Errorf("generate slug: %w", err)
	}

	album, err := s.queries.CreatePhotoAlbum(ctx, sqlc.CreatePhotoAlbumParams{
		Slug:        slug,
		UserID:      userID,
		Title:       req.Title,
		Description: req.Description,
	})
	if err != nil {
		return nil, fmt.Errorf("create album: %w", err)
	}

	return albumToResponse(album), nil
}

func (s *PhotoService) ListAlbums(ctx context.Context, userID int64, page, pageSize int) (*AlbumListResponse, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 50
	}

	total, err := s.queries.CountPhotoAlbumsByUser(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("count albums: %w", err)
	}

	albums, err := s.queries.ListPhotoAlbumsByUser(ctx, sqlc.ListPhotoAlbumsByUserParams{
		UserID: userID,
		Limit:  int32(pageSize),
		Offset: int32((page - 1) * pageSize),
	})
	if err != nil {
		return nil, fmt.Errorf("list albums: %w", err)
	}

	items := make([]AlbumResponse, 0, len(albums))
	for _, a := range albums {
		items = append(items, *albumToResponse(a))
	}

	return &AlbumListResponse{Items: items, Total: int(total)}, nil
}

func (s *PhotoService) GetAlbum(ctx context.Context, userID int64, slug string) (*AlbumResponse, error) {
	album, err := s.queries.GetPhotoAlbumBySlug(ctx, sqlc.GetPhotoAlbumBySlugParams{
		Slug:   slug,
		UserID: userID,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, model.ErrNotFound
		}
		return nil, fmt.Errorf("get album: %w", err)
	}

	return albumToResponse(album), nil
}

func (s *PhotoService) UpdateAlbum(ctx context.Context, userID int64, slug string, req AlbumUpdateRequest) (*AlbumResponse, error) {
	cover := pgtype.Text{Valid: false}
	if req.CoverFileSlug != nil {
		cover = pgtype.Text{String: *req.CoverFileSlug, Valid: true}
	}

	album, err := s.queries.UpdatePhotoAlbum(ctx, sqlc.UpdatePhotoAlbumParams{
		Slug:          slug,
		UserID:        userID,
		Title:         req.Title,
		Description:   req.Description,
		CoverFileSlug: cover,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, model.ErrNotFound
		}
		return nil, fmt.Errorf("update album: %w", err)
	}

	return albumToResponse(album), nil
}

func (s *PhotoService) DeleteAlbum(ctx context.Context, userID int64, slug string) error {
	result, err := s.pg.Exec(ctx, "DELETE FROM photo_albums WHERE slug = $1 AND user_id = $2", slug, userID)
	if err != nil {
		return fmt.Errorf("delete album: %w", err)
	}
	if result.RowsAffected() == 0 {
		return model.ErrNotFound
	}
	return nil
}

type AlbumItemResponse struct {
	ID          int64   `json:"id"`
	FileSlug    string  `json:"fileSlug"`
	FileName    string  `json:"fileName"`
	FileSize    int64   `json:"fileSize"`
	MimeType    *string `json:"mimeType"`
	CreatedAt   string  `json:"createdAt"`
	ThumbnailURL string `json:"thumbnailUrl"`
	FileHash    *string `json:"fileHash"`
}

func (s *PhotoService) AddPhotosToAlbum(ctx context.Context, userID int64, albumSlug string, fileSlugs []string) error {
	album, err := s.queries.GetPhotoAlbumBySlug(ctx, sqlc.GetPhotoAlbumBySlugParams{
		Slug:   albumSlug,
		UserID: userID,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return model.ErrNotFound
		}
		return fmt.Errorf("get album: %w", err)
	}

	for i, fs := range fileSlugs {
		_, err := s.queries.AddPhotoToAlbum(ctx, sqlc.AddPhotoToAlbumParams{
			AlbumID:   album.ID,
			FileSlug:  fs,
			SortOrder: int32(i),
		})
		if err != nil {
			if isPGUniqueViolation(err) {
				continue
			}
			return fmt.Errorf("add photo to album: %w", err)
		}
		_ = s.queries.IncrementPhotoAlbumItemCount(ctx, album.ID)
	}

	return nil
}

func (s *PhotoService) RemovePhotoFromAlbum(ctx context.Context, userID int64, albumSlug, fileSlug string) error {
	album, err := s.queries.GetPhotoAlbumBySlug(ctx, sqlc.GetPhotoAlbumBySlugParams{
		Slug:   albumSlug,
		UserID: userID,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return model.ErrNotFound
		}
		return fmt.Errorf("get album: %w", err)
	}

	if err := s.queries.RemovePhotoFromAlbum(ctx, sqlc.RemovePhotoFromAlbumParams{
		AlbumID:  album.ID,
		FileSlug: fileSlug,
	}); err != nil {
		return fmt.Errorf("remove photo: %w", err)
	}

	_ = s.queries.DecrementPhotoAlbumItemCount(ctx, album.ID)
	return nil
}

func (s *PhotoService) ListAlbumPhotos(ctx context.Context, userID int64, albumSlug string, page, pageSize int) (*PhotoListResponse, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 50
	}

	album, err := s.queries.GetPhotoAlbumBySlug(ctx, sqlc.GetPhotoAlbumBySlugParams{
		Slug:   albumSlug,
		UserID: userID,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, model.ErrNotFound
		}
		return nil, fmt.Errorf("get album: %w", err)
	}

	total, err := s.queries.CountPhotoAlbumItems(ctx, album.ID)
	if err != nil {
		return nil, fmt.Errorf("count album items: %w", err)
	}

	rows, err := s.queries.ListPhotoAlbumItems(ctx, sqlc.ListPhotoAlbumItemsParams{
		AlbumID: album.ID,
		Limit:   int32(pageSize),
		Offset:  int32((page - 1) * pageSize),
	})
	if err != nil {
		return nil, fmt.Errorf("list album items: %w", err)
	}

	items := make([]PhotoItem, 0, len(rows))
	for _, r := range rows {
		item := PhotoItem{
			Slug:     r.FileSlug,
			FileName: r.FileName,
			FileSize: r.FileSize,
			CreatedAt: r.FileCreatedAt.Time.Format("2006-01-02T15:04:05Z"),
		}
		if r.MimeType.Valid {
			m := r.MimeType.String
			item.MimeType = &m
		}
		if r.FileHash.Valid {
			h := r.FileHash.String
			item.FileHash = &h
			item.ThumbnailURL = fmt.Sprintf("/api/v1/photos/%s/thumbnail", r.FileSlug)
		}
		items = append(items, item)
	}

	return &PhotoListResponse{Items: items, Total: int(total), Page: page}, nil
}

func (s *PhotoService) ListPhotoAlbums(ctx context.Context, userID int64, fileSlug string) ([]AlbumResponse, error) {
	albums, err := s.queries.ListPhotoAlbumsByFile(ctx, sqlc.ListPhotoAlbumsByFileParams{
		FileSlug: fileSlug,
		UserID:   userID,
	})
	if err != nil {
		return nil, fmt.Errorf("list photo albums: %w", err)
	}

	items := make([]AlbumResponse, 0, len(albums))
	for _, a := range albums {
		items = append(items, *albumToResponse(a))
	}
	return items, nil
}

func albumToResponse(a sqlc.PhotoAlbum) *AlbumResponse {
	resp := &AlbumResponse{
		Slug:        a.Slug,
		Title:       a.Title,
		Description: a.Description,
		ItemCount:   int(a.ItemCount),
		CreatedAt:   a.CreatedAt.Time.Format("2006-01-02T15:04:05Z"),
		UpdatedAt:   a.UpdatedAt.Time.Format("2006-01-02T15:04:05Z"),
	}
	if a.CoverFileSlug.Valid {
		u := fmt.Sprintf("/api/v1/photos/%s/thumbnail", a.CoverFileSlug.String)
		resp.CoverURL = &u
	}
	return resp
}

func isPGUniqueViolation(err error) bool {
	return err != nil && stringsContains(err.Error(), "unique")
}

func stringsContains(s, substr string) bool {
	return len(s) >= len(substr) && containsStr(s, substr)
}

func containsStr(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// Ensure Math import is used
var _ = math.MaxInt64
