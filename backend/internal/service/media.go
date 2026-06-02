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
	files   *FilesService
	logger  zerolog.Logger
}

func NewMediaService(
	queries *sqlc.Queries,
	pg *pgxpool.Pool,
	cfg *config.Config,
	store *storage.Local,
	c *cache.Cache,
	files *FilesService,
	logger zerolog.Logger,
) *MediaService {
	return &MediaService{
		queries: queries,
		pg:      pg,
		cfg:     cfg,
		store:   store,
		cache:   c,
		files:   files,
		logger:  logger,
	}
}

type AddToLibraryRequest struct {
	FileSlug string `json:"fileSlug"`
	Title    string `json:"title"`
}

type AddToLibraryResponse struct {
	MediaSlug       string `json:"mediaSlug"`
	TranscodeSlug   string `json:"transcodeSlug"`
	TranscodeStatus string `json:"transcodeStatus"`
	TranscodeReused bool   `json:"transcodeReused"`
}

type MediaItemResponse struct {
	MediaSlug       string  `json:"mediaSlug"`
	FileName        string  `json:"fileName"`
	Status          string  `json:"status"`
	TranscodeReused bool    `json:"transcodeReused"`
	Progress        int32   `json:"progress"`
	DurationSec     *int32  `json:"durationSec"`
	ErrorMsg        *string `json:"errorMsg"`
	PosterURL       *string `json:"posterUrl"`
	PlayURL         *string `json:"playUrl"`
	CreatedAt       string  `json:"createdAt"`
}

func (s *MediaService) EnsureUploadDir(ctx context.Context, userID int64) (*FileItem, error) {
	return s.files.EnsureSystemDir(ctx, userID, SystemDirOptions{
		Kind: SystemDirMediaUploads,
		Name: "媒体库上传",
	})
}

func (s *MediaService) AddToLibrary(ctx context.Context, userID int64, req AddToLibraryRequest) (*AddToLibraryResponse, error) {
	log := s.logger.With().Int64("user_id", userID).Str("file_slug", req.FileSlug).Logger()

	if req.FileSlug == "" {
		log.Warn().Msg("add to library: empty file slug")
		return nil, model.ErrInvalidInput
	}
	log.Info().Msg("add to library")

	// Get the user file
	uf, err := s.queries.GetFileBySlugForUser(ctx, sqlc.GetFileBySlugForUserParams{
		Slug:   req.FileSlug,
		UserID: userID,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			log.Warn().Msg("add to library: file not found")
			return nil, model.ErrNotFound
		}
		return nil, fmt.Errorf("get file: %w", err)
	}
	log.Debug().Str("file_name", uf.FileName).Msg("file loaded")

	if uf.IsDir || uf.IsTrashed || !uf.PhysicalFileID.Valid {
		log.Warn().Bool("is_dir", uf.IsDir).Bool("is_trashed", uf.IsTrashed).Bool("has_physical_file", uf.PhysicalFileID.Valid).Msg("add to library: invalid file state")
		return nil, model.ErrInvalidInput
	}

	// Check if already in library
	existing, err := s.queries.GetMediaItemByUserAndFile(ctx, sqlc.GetMediaItemByUserAndFileParams{
		UserID:     userID,
		UserFileID: uf.ID,
	})
	if err == nil && existing.ID != 0 {
		status := "unknown"
		if existing.TranscodeID.Valid {
			tc, tcErr := s.queries.GetTranscodeByID(ctx, existing.TranscodeID.Int64)
			if tcErr == nil {
				status = tc.Status
			}
		}
		log.Info().Str("media_slug", existing.Slug).Str("transcode_status", status).Msg("media already in library")
		return &AddToLibraryResponse{
			MediaSlug:       existing.Slug,
			TranscodeStatus: status,
		}, nil
	}

	// Find or create shared transcode
	profile := "hls_1080p"
	tc, err := s.queries.GetTranscodeByPhysicalFileAndProfile(ctx, sqlc.GetTranscodeByPhysicalFileAndProfileParams{
		PhysicalFileID: uf.PhysicalFileID.Int64,
		Profile:        profile,
	})

	transcodeReused := false
	var transcodeSlug string

	if err != nil && errors.Is(err, pgx.ErrNoRows) {
		log.Debug().Msg("creating new transcode")
		tc, transcodeSlug, err = s.createTranscodeWithJob(ctx, uf.PhysicalFileID.Int64, profile)
		if err != nil {
			return nil, err
		}
	} else if err != nil {
		return nil, fmt.Errorf("get transcode: %w", err)
	} else if tc.Status == "failed" {
		log.Debug().Str("old_transcode_slug", tc.Slug).Msg("recreating failed transcode")
		tc, transcodeSlug, err = s.createTranscodeWithJob(ctx, uf.PhysicalFileID.Int64, profile)
		if err != nil {
			return nil, err
		}
	} else {
		transcodeReused = true
		transcodeSlug = tc.Slug
		log.Debug().Str("transcode_slug", tc.Slug).Str("status", tc.Status).Msg("existing transcode reused")
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

	log.Info().Str("media_slug", mediaSlug).Str("transcode_slug", transcodeSlug).Bool("reused", transcodeReused).Msg("media added to library")

	return &AddToLibraryResponse{
		MediaSlug:       mediaSlug,
		TranscodeSlug:   transcodeSlug,
		TranscodeStatus: tc.Status,
		TranscodeReused: transcodeReused,
	}, nil
}

func (s *MediaService) createTranscodeWithJob(ctx context.Context, physicalFileID int64, profile string) (sqlc.MediaTranscode, string, error) {
	tcSlug, err := gonanoid.New(21)
	if err != nil {
		return sqlc.MediaTranscode{}, "", fmt.Errorf("generate slug: %w", err)
	}

	tc, err := s.queries.CreateTranscode(ctx, sqlc.CreateTranscodeParams{
		Slug:           tcSlug,
		PhysicalFileID: physicalFileID,
		Profile:        profile,
		Status:         "pending",
	})
	if err != nil {
		return sqlc.MediaTranscode{}, "", fmt.Errorf("create transcode: %w", err)
	}

	jobSlug, err := gonanoid.New(21)
	if err != nil {
		return sqlc.MediaTranscode{}, "", fmt.Errorf("generate slug: %w", err)
	}
	_, err = s.queries.CreateMediaJob(ctx, sqlc.CreateMediaJobParams{
		Slug:        jobSlug,
		TranscodeID: tc.ID,
		Status:      "pending",
	})
	if err != nil {
		return sqlc.MediaTranscode{}, "", fmt.Errorf("create media job: %w", err)
	}

	return tc, tcSlug, nil
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
			MediaSlug: item.Slug,
			FileName:  item.Title,
			CreatedAt: item.CreatedAt.Time.Format("2006-01-02T15:04:05Z"),
		}
		if item.TranscodeStatus.Valid {
			r.Status = item.TranscodeStatus.String
		}
		if item.DurationSec.Valid {
			r.DurationSec = &item.DurationSec.Int32
		}
		if r.Status == "processing" && item.TranscodeSlug.Valid {
			progress, _ := s.cache.MediaProgress.GetProgress(ctx, item.TranscodeSlug.String)
			r.Progress = int32(progress)
		}
		if item.PosterPath.Valid && item.PosterPath.String != "" {
			posterURL := fmt.Sprintf("/api/v1/media/poster/%s", item.Slug)
			r.PosterURL = &posterURL
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
		MediaSlug: item.Slug,
		FileName:  item.Title,
		CreatedAt: item.CreatedAt.Time.Format("2006-01-02T15:04:05Z"),
	}

	if item.TranscodeStatus.Valid {
		resp.Status = item.TranscodeStatus.String
	}
	if item.DurationSec.Valid {
		resp.DurationSec = &item.DurationSec.Int32
	}

	// Return progress from Redis if processing
	if resp.Status == "processing" && item.TranscodeSlug.Valid {
		progress, _ := s.cache.MediaProgress.GetProgress(ctx, item.TranscodeSlug.String)
		resp.Progress = int32(progress)
	}

	// Set play URL if done
	if resp.Status == "done" && item.TranscodeSlug.Valid {
		playURL := fmt.Sprintf("/api/v1/media/hls/%s/index.m3u8", item.Slug)
		resp.PlayURL = &playURL
	}

	// Set poster URL
	if item.PosterPath.Valid && item.PosterPath.String != "" {
		posterURL := fmt.Sprintf("/api/v1/media/poster/%s", item.Slug)
		resp.PosterURL = &posterURL
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

func (s *MediaService) GetPosterPath(ctx context.Context, userID int64, mediaSlug string) (string, error) {
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
	if !item.PosterPath.Valid || item.PosterPath.String == "" {
		return "", model.ErrNotFound
	}
	return item.PosterPath.String, nil
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
