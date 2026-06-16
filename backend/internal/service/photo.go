package service

import (
	"context"
	"errors"
	"fmt"
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
	Slug         string  `json:"slug"`
	FileName     string  `json:"fileName"`
	FileSize     int64   `json:"fileSize"`
	MimeType     *string `json:"mimeType"`
	IsStarred    bool    `json:"isStarred"`
	CreatedAt    string  `json:"createdAt"`
	UpdatedAt    string  `json:"updatedAt"`
	ThumbnailURL string  `json:"thumbnailUrl"`
	FileHash     *string `json:"fileHash"`
	ParentSlug   *string `json:"parentSlug"`
}

type PhotoListResponse struct {
	Items []PhotoItem `json:"items"`
	Total int         `json:"total"`
	Page  int         `json:"page"`
}

func (s *PhotoService) ListPhotos(ctx context.Context, userID int64, sessionID string, page, pageSize int) (*PhotoListResponse, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 50
	}

	// Phase 1: fetch all candidate IDs ordered by created_at DESC.
	idRows, err := s.pg.Query(ctx, `
		SELECT uf.id FROM user_files uf
		WHERE uf.user_id = $1
		  AND uf.is_dir = FALSE
		  AND uf.is_trashed = FALSE
		  AND uf.file_category = 'image'
		ORDER BY uf.created_at DESC
	`, userID)
	if err != nil {
		return nil, fmt.Errorf("list image ids: %w", err)
	}
	var allIDs []int64
	for idRows.Next() {
		var id int64
		if err := idRows.Scan(&id); err != nil {
			idRows.Close()
			return nil, fmt.Errorf("scan image id: %w", err)
		}
		allIDs = append(allIDs, id)
	}
	idRows.Close()
	if err := idRows.Err(); err != nil {
		return nil, err
	}

	// Phase 2: filter out files hidden by locked directories.
	visibleIDs, err := filterLockedFileIDs(ctx, s.pg, userID, sessionID, allIDs)
	if err != nil {
		return nil, fmt.Errorf("filter locked photos: %w", err)
	}
	total := len(visibleIDs)

	// Phase 3: paginate the visible IDs.
	start := (page - 1) * pageSize
	if start >= len(visibleIDs) {
		return &PhotoListResponse{Items: []PhotoItem{}, Total: total, Page: page}, nil
	}
	end := start + pageSize
	if end > len(visibleIDs) {
		end = len(visibleIDs)
	}
	pageIDs := visibleIDs[start:end]

	if len(pageIDs) == 0 {
		return &PhotoListResponse{Items: []PhotoItem{}, Total: total, Page: page}, nil
	}

	// Phase 4: fetch full row data for only the visible page IDs, preserving order.
	items, err := s.fetchImageRows(ctx, pageIDs)
	if err != nil {
		return nil, err
	}

	return &PhotoListResponse{Items: items, Total: total, Page: page}, nil
}

// fetchImageRows fetches full image row data for the given file IDs, preserving the input order.
func (s *PhotoService) fetchImageRows(ctx context.Context, ids []int64) ([]PhotoItem, error) {
	rows, err := s.pg.Query(ctx, `
		SELECT uf.id, uf.slug, uf.file_name, uf.file_size, uf.mime_type,
		       uf.is_starred, uf.created_at, uf.updated_at, pf.file_hash, uf.parent_slug
		FROM user_files uf
		LEFT JOIN physical_files pf ON pf.id = uf.physical_file_id
		WHERE uf.id = ANY($1::bigint[])
	`, ids)
	if err != nil {
		return nil, fmt.Errorf("fetch image rows: %w", err)
	}
	defer rows.Close()

	type imageRow struct {
		ID         int64
		Slug       string
		FileName   string
		FileSize   int64
		MimeType   pgtype.Text
		IsStarred  bool
		CreatedAt  pgtype.Timestamptz
		UpdatedAt  pgtype.Timestamptz
		FileHash   pgtype.Text
		ParentSlug pgtype.Text
	}
	rowMap := make(map[int64]imageRow, len(ids))
	for rows.Next() {
		var r imageRow
		if err := rows.Scan(&r.ID, &r.Slug, &r.FileName, &r.FileSize, &r.MimeType,
			&r.IsStarred, &r.CreatedAt, &r.UpdatedAt, &r.FileHash, &r.ParentSlug); err != nil {
			return nil, fmt.Errorf("scan image row: %w", err)
		}
		rowMap[r.ID] = r
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	items := make([]PhotoItem, 0, len(ids))
	for _, id := range ids {
		r, ok := rowMap[id]
		if !ok {
			continue
		}
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
		if r.ParentSlug.Valid && r.ParentSlug.String != "" {
			ps := r.ParentSlug.String
			item.ParentSlug = &ps
		}
		items = append(items, item)
	}
	return items, nil
}

type ThumbnailResult struct {
	Path     string
	FileHash string
}

func (s *PhotoService) GetThumbnailPath(ctx context.Context, userID int64, sessionID, fileSlug string) (*ThumbnailResult, error) {
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

	if !uf.PhysicalFileID.Valid {
		return nil, model.ErrNotFound
	}
	if err := ensureFileUnlocked(ctx, s.pg, userID, sessionID, uf.ID); err != nil {
		return nil, err
	}

	pf, err := s.queries.GetPhysicalFileByID(ctx, uf.PhysicalFileID.Int64)
	if err != nil {
		return nil, fmt.Errorf("get physical file: %w", err)
	}

	thumbPath, err := storage.ServeThumbnail(s.cfg.Storage.Root, s.cfg.Storage.FilesDir, pf.FileHash)
	if err != nil {
		return nil, fmt.Errorf("serve thumbnail: %w", err)
	}

	return &ThumbnailResult{Path: thumbPath, FileHash: pf.FileHash}, nil
}

func (s *PhotoService) GetPhotoDetail(ctx context.Context, userID int64, sessionID, fileSlug string) (*PhotoItem, error) {
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
	if err := ensureFileUnlocked(ctx, s.pg, userID, sessionID, uf.ID); err != nil {
		return nil, err
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
	ID           int64   `json:"id"`
	FileSlug     string  `json:"fileSlug"`
	FileName     string  `json:"fileName"`
	FileSize     int64   `json:"fileSize"`
	MimeType     *string `json:"mimeType"`
	CreatedAt    string  `json:"createdAt"`
	ThumbnailURL string  `json:"thumbnailUrl"`
	FileHash     *string `json:"fileHash"`
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

func (s *PhotoService) ListAlbumPhotos(ctx context.Context, userID int64, sessionID, albumSlug string, page, pageSize int) (*PhotoListResponse, error) {
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

	// Phase 1: fetch all candidate file IDs ordered by sort_order, added_at.
	idRows, err := s.pg.Query(ctx, `
		SELECT uf.id FROM photo_album_items pai
		JOIN user_files uf ON uf.slug = pai.file_slug
		WHERE pai.album_id = $1
		ORDER BY pai.sort_order ASC, pai.added_at DESC
	`, album.ID)
	if err != nil {
		return nil, fmt.Errorf("list album item ids: %w", err)
	}
	var allIDs []int64
	for idRows.Next() {
		var id int64
		if err := idRows.Scan(&id); err != nil {
			idRows.Close()
			return nil, fmt.Errorf("scan album item id: %w", err)
		}
		allIDs = append(allIDs, id)
	}
	idRows.Close()
	if err := idRows.Err(); err != nil {
		return nil, err
	}

	// Phase 2: filter out files hidden by locked directories.
	visibleIDs, err := filterLockedFileIDs(ctx, s.pg, userID, sessionID, allIDs)
	if err != nil {
		return nil, fmt.Errorf("filter locked album photos: %w", err)
	}
	total := len(visibleIDs)

	// Phase 3: paginate the visible IDs.
	start := (page - 1) * pageSize
	if start >= len(visibleIDs) {
		return &PhotoListResponse{Items: []PhotoItem{}, Total: total, Page: page}, nil
	}
	end := start + pageSize
	if end > len(visibleIDs) {
		end = len(visibleIDs)
	}
	pageIDs := visibleIDs[start:end]

	if len(pageIDs) == 0 {
		return &PhotoListResponse{Items: []PhotoItem{}, Total: total, Page: page}, nil
	}

	// Phase 4: fetch full row data for only the visible page IDs, preserving order.
	items, err := s.fetchAlbumItemRows(ctx, pageIDs)
	if err != nil {
		return nil, err
	}

	return &PhotoListResponse{Items: items, Total: total, Page: page}, nil
}

// fetchAlbumItemRows fetches full album item row data for the given file IDs, preserving input order.
func (s *PhotoService) fetchAlbumItemRows(ctx context.Context, ids []int64) ([]PhotoItem, error) {
	rows, err := s.pg.Query(ctx, `
		SELECT uf.id, uf.slug, uf.file_name, uf.file_size, uf.mime_type,
		       uf.created_at, pf.file_hash, uf.parent_slug
		FROM user_files uf
		LEFT JOIN physical_files pf ON pf.id = uf.physical_file_id
		WHERE uf.id = ANY($1::bigint[])
	`, ids)
	if err != nil {
		return nil, fmt.Errorf("fetch album item rows: %w", err)
	}
	defer rows.Close()

	type albumItemRow struct {
		FileID        int64
		FileSlug      string
		FileName      string
		FileSize      int64
		MimeType      pgtype.Text
		FileCreatedAt pgtype.Timestamptz
		FileHash      pgtype.Text
		ParentSlug    pgtype.Text
	}
	rowMap := make(map[int64]albumItemRow, len(ids))
	for rows.Next() {
		var r albumItemRow
		if err := rows.Scan(&r.FileID, &r.FileSlug, &r.FileName, &r.FileSize, &r.MimeType,
			&r.FileCreatedAt, &r.FileHash, &r.ParentSlug); err != nil {
			return nil, fmt.Errorf("scan album item row: %w", err)
		}
		rowMap[r.FileID] = r
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	items := make([]PhotoItem, 0, len(ids))
	for _, id := range ids {
		r, ok := rowMap[id]
		if !ok {
			continue
		}
		item := PhotoItem{
			Slug:      r.FileSlug,
			FileName:  r.FileName,
			FileSize:  r.FileSize,
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
		if r.ParentSlug.Valid && r.ParentSlug.String != "" {
			ps := r.ParentSlug.String
			item.ParentSlug = &ps
		}
		items = append(items, item)
	}
	return items, nil
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
