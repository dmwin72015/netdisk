package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	gonanoid "github.com/matoous/go-nanoid/v2"

	"github.com/netdisk/server/internal/cache"
	"github.com/netdisk/server/internal/config"
	"github.com/netdisk/server/internal/db/sqlc"
	"github.com/netdisk/server/internal/model"
	"github.com/netdisk/server/internal/storage"
)

type MediaService struct {
	queries *sqlc.Queries
	pg      *pgxpool.Pool
	cfg     *config.Config
	store   *storage.Local
	cache   *cache.Cache
}

func NewMediaService(queries *sqlc.Queries, pg *pgxpool.Pool, cfg *config.Config, store *storage.Local, c *cache.Cache) *MediaService {
	return &MediaService{queries: queries, pg: pg, cfg: cfg, store: store, cache: c}
}

type AddToLibraryRequest struct {
	FileSlug string `json:"file_slug"`
	Title    string `json:"title"`
}

type AddToLibraryResponse struct {
	MediaSlug       string `json:"media_slug"`
	TranscodeSlug   string `json:"transcode_slug"`
	TranscodeStatus string `json:"transcode_status"`
	TranscodeReused bool   `json:"transcode_reused"`
}

type MediaItemResponse struct {
	Slug            string  `json:"slug"`
	Title           string  `json:"title"`
	TranscodeStatus string  `json:"transcode_status"`
	TranscodeReused bool    `json:"transcode_reused"`
	DurationSec     *int32  `json:"duration_sec"`
	PosterURL       *string `json:"poster_url"`
	PlayURL         *string `json:"play_url"`
	CreatedAt       string  `json:"created_at"`
}

func (s *MediaService) AddToLibrary(ctx context.Context, userID int64, req AddToLibraryRequest) (*AddToLibraryResponse, error) {
	if req.FileSlug == "" {
		return nil, model.ErrInvalidInput
	}

	// Get the user file
	uf, err := s.queries.GetFileBySlugForUser(ctx, sqlc.GetFileBySlugForUserParams{
		Slug:   req.FileSlug,
		UserID: userID,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, model.ErrNotFound
		}
		return nil, fmt.Errorf("get file: %w", err)
	}
	if uf.IsDir || uf.IsTrashed || !uf.PhysicalFileID.Valid {
		return nil, model.ErrInvalidInput
	}

	// Check if already in library
	existing, err := s.queries.GetMediaItemByUserAndFile(ctx, sqlc.GetMediaItemByUserAndFileParams{
		UserID:     userID,
		UserFileID: uf.ID,
	})
	if err == nil && existing.ID != 0 {
		// Already in library, get transcode info
		tc, _ := s.queries.GetTranscodeBySlug(ctx, "")
		if existing.TranscodeID.Valid {
			tc, _ = s.queries.GetTranscodeBySlug(ctx, "")
			// Get transcode by ID
		}
		_ = tc
		return &AddToLibraryResponse{
			MediaSlug:       existing.Slug,
			TranscodeStatus: "existing",
		}, nil
	}

	// Find or create shared transcode
	profile := "default"
	tc, err := s.queries.GetTranscodeByPhysicalFileAndProfile(ctx, sqlc.GetTranscodeByPhysicalFileAndProfileParams{
		PhysicalFileID: uf.PhysicalFileID.Int64,
		Profile:        profile,
	})

	transcodeReused := false
	var transcodeSlug string

	if err != nil && errors.Is(err, pgx.ErrNoRows) {
		// Create new transcode
		tcSlug, err := gonanoid.New(21)
		if err != nil {
			return nil, fmt.Errorf("generate slug: %w", err)
		}

		tc, err = s.queries.CreateTranscode(ctx, sqlc.CreateTranscodeParams{
			Slug:           tcSlug,
			PhysicalFileID: uf.PhysicalFileID.Int64,
			Profile:        profile,
			Status:         "pending",
		})
		if err != nil {
			return nil, fmt.Errorf("create transcode: %w", err)
		}

		// Create media job
		jobSlug, err := gonanoid.New(21)
		if err != nil {
			return nil, fmt.Errorf("generate slug: %w", err)
		}
		_, err = s.queries.CreateMediaJob(ctx, sqlc.CreateMediaJobParams{
			Slug:        jobSlug,
			TranscodeID: tc.ID,
			Status:      "pending",
		})
		if err != nil {
			return nil, fmt.Errorf("create media job: %w", err)
		}

		transcodeSlug = tc.Slug
	} else if err != nil {
		return nil, fmt.Errorf("get transcode: %w", err)
	} else {
		// Reuse existing transcode
		transcodeReused = true
		transcodeSlug = tc.Slug
	}

	// Create media item
	title := req.Title
	if title == "" {
		title = uf.FileName
	}

	mediaSlug, err := gonanoid.New(21)
	if err != nil {
		return nil, fmt.Errorf("generate slug: %w", err)
	}

	_, err = s.queries.CreateMediaItem(ctx, sqlc.CreateMediaItemParams{
		Slug:           mediaSlug,
		UserID:         userID,
		UserFileID:     uf.ID,
		PhysicalFileID: uf.PhysicalFileID.Int64,
		TranscodeID:    pgtype.Int8{Int64: tc.ID, Valid: true},
		Title:          title,
	})
	if err != nil {
		return nil, fmt.Errorf("create media item: %w", err)
	}

	return &AddToLibraryResponse{
		MediaSlug:       mediaSlug,
		TranscodeSlug:   transcodeSlug,
		TranscodeStatus: tc.Status,
		TranscodeReused: transcodeReused,
	}, nil
}

func (s *MediaService) ListMediaItems(ctx context.Context, userID int64, page, pageSize int) ([]MediaItemResponse, int, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	total, err := s.queries.CountMediaItemsByUser(ctx, userID)
	if err != nil {
		return nil, 0, fmt.Errorf("count media items: %w", err)
	}

	items, err := s.queries.ListMediaItemsByUser(ctx, sqlc.ListMediaItemsByUserParams{
		UserID: userID,
		Limit:  int32(pageSize),
		Offset: int32((page - 1) * pageSize),
	})
	if err != nil {
		return nil, 0, fmt.Errorf("list media items: %w", err)
	}

	resp := make([]MediaItemResponse, 0, len(items))
	for _, item := range items {
		r := MediaItemResponse{
			Slug:      item.Slug,
			Title:     item.Title,
			CreatedAt: item.CreatedAt.Time.Format("2006-01-02T15:04:05Z"),
		}
		if item.TranscodeStatus.Valid {
			r.TranscodeStatus = item.TranscodeStatus.String
		}
		if item.DurationSec.Valid {
			r.DurationSec = &item.DurationSec.Int32
		}
		resp = append(resp, r)
	}

	return resp, int(total), nil
}

func (s *MediaService) GetMediaItem(ctx context.Context, userID int64, mediaSlug string) (*MediaItemResponse, error) {
	item, err := s.queries.GetMediaItemBySlug(ctx, mediaSlug)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, model.ErrNotFound
		}
		return nil, fmt.Errorf("get media item: %w", err)
	}
	if item.UserID != userID {
		return nil, model.ErrNotFound
	}

	resp := &MediaItemResponse{
		Slug:      item.Slug,
		Title:     item.Title,
		CreatedAt: item.CreatedAt.Time.Format("2006-01-02T15:04:05Z"),
	}

	if item.TranscodeStatus.Valid {
		resp.TranscodeStatus = item.TranscodeStatus.String
	}
	if item.DurationSec.Valid {
		resp.DurationSec = &item.DurationSec.Int32
	}

	// Check Redis for progress if processing
	if resp.TranscodeStatus == "processing" && item.TranscodeSlug.Valid {
		progress, _ := s.cache.MediaProgress.GetProgress(ctx, item.TranscodeSlug.String)
		if progress > 0 {
			_ = progress
		}
	}

	// Set play URL if done
	if resp.TranscodeStatus == "done" && item.TranscodeSlug.Valid {
		playURL := fmt.Sprintf("/api/v1/media/hls/%s/index.m3u8", item.Slug)
		resp.PlayURL = &playURL
	}

	return resp, nil
}

func (s *MediaService) RemoveFromLibrary(ctx context.Context, userID int64, mediaSlug string) error {
	item, err := s.queries.GetMediaItemBySlug(ctx, mediaSlug)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return model.ErrNotFound
		}
		return fmt.Errorf("get media item: %w", err)
	}
	if item.UserID != userID {
		return model.ErrNotFound
	}

	return s.queries.DeleteMediaItem(ctx, item.ID)
}

func (s *MediaService) GetHLSPath(ctx context.Context, userID int64, mediaSlug, subPath string) (string, error) {
	item, err := s.queries.GetMediaItemBySlug(ctx, mediaSlug)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return "", model.ErrNotFound
		}
		return "", fmt.Errorf("get media item: %w", err)
	}
	if item.UserID != userID {
		return "", model.ErrNotFound
	}

	if !item.TranscodeStatus.Valid || item.TranscodeStatus.String != "done" {
		return "", model.ErrNotFound
	}
	if !item.HlsDir.Valid {
		return "", model.ErrNotFound
	}

	// Validate subPath doesn't escape HLS dir
	cleanPath := storage.SafePath(subPath)
	if cleanPath == "" {
		cleanPath = "index.m3u8"
	}

	fullPath := fmt.Sprintf("%s/%s", item.HlsDir.String, cleanPath)
	return fullPath, nil
}
