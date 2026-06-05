package media

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/zerolog"

	"github.com/netdisk/server/internal/cache"
	"github.com/netdisk/server/internal/config"
	"github.com/netdisk/server/internal/db/sqlc"
	"github.com/netdisk/server/internal/storage"
)

type Worker struct {
	queries     *sqlc.Queries
	pg          *pgxpool.Pool
	cfg         *config.Config
	store       *storage.Local
	cache       *cache.Cache
	logger      zerolog.Logger
	broadcaster *Broadcaster
}

func NewWorker(queries *sqlc.Queries, pg *pgxpool.Pool, cfg *config.Config, store *storage.Local, c *cache.Cache, logger zerolog.Logger, broadcaster *Broadcaster) *Worker {
	return &Worker{
		queries:     queries,
		pg:          pg,
		cfg:         cfg,
		store:       store,
		cache:       c,
		logger:      logger,
		broadcaster: broadcaster,
	}
}

// Start begins the polling loop for pending media jobs.
func (w *Worker) Start(ctx context.Context) {
	w.logger.Info().Msg("media worker started")

	ticker := time.NewTicker(w.cfg.Media.PollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			w.logger.Info().Msg("media worker stopped")
			return
		case <-ticker.C:
			w.processJobs(ctx)
		}
	}
}

func (w *Worker) processJobs(ctx context.Context) {
	jobs, err := w.queries.GetPendingJobs(ctx, int32(w.cfg.Media.BatchSize))
	if err != nil {
		w.logger.Error().Err(err).Msg("get pending jobs")
		return
	}

	for _, job := range jobs {
		w.processJob(ctx, job)
	}
}

func (w *Worker) processJob(ctx context.Context, job sqlc.MediaJob) {
	log := w.logger.With().Str("job_slug", job.Slug).Int64("job_id", job.ID).Logger()
	log.Info().Msg("processing job")

	// Mark job as processing
	if err := w.queries.UpdateJobStatus(ctx, sqlc.UpdateJobStatusParams{
		ID:     job.ID,
		Status: "processing",
	}); err != nil {
		log.Error().Err(err).Msg("update job status to processing")
		return
	}
	log.Debug().Msg("job status set to processing")

	// Get transcode info
	tc, err := w.queries.GetTranscodeByID(ctx, job.TranscodeID)
	if err != nil {
		log.Error().Err(err).Int64("transcode_id", job.TranscodeID).Msg("get transcode")
		w.failJob(ctx, job.ID, "transcode not found")
		return
	}
	log.Debug().Str("transcode_slug", tc.Slug).Str("profile", tc.Profile).Msg("transcode loaded")

	// Mark transcode as processing
	if err := w.queries.UpdateTranscodeStatus(ctx, sqlc.UpdateTranscodeStatusParams{
		ID:     tc.ID,
		Status: "processing",
	}); err != nil {
		log.Error().Err(err).Msg("update transcode status to processing")
		return
	}

	// Look up media item for SSE progress events
	var mediaSlug string
	var mediaUserID int64
	mi, miErr := w.queries.GetMediaItemByTranscodeID(ctx, pgtype.Int8{Int64: tc.ID, Valid: true})
	if miErr == nil {
		mediaSlug = mi.Slug
		mediaUserID = mi.UserID
	}

	// Get physical file
	pf, err := w.queries.GetPhysicalFileByID(ctx, tc.PhysicalFileID)
	if err != nil {
		log.Error().Err(err).Int64("physical_file_id", tc.PhysicalFileID).Msg("get physical file")
		w.publishFailEvent(mediaSlug, mediaUserID, "physical file not found")
		w.failTranscodeAndJob(ctx, tc.ID, job.ID, "physical file not found")
		return
	}
	log.Debug().Str("file_hash", pf.FileHash).Int64("file_size", pf.FileSize).Msg("physical file loaded")

	// Build paths
	inputPath := w.store.AbsPath(pf.FileHash)
	outputDir := storage.HLSAbsPath(w.cfg.Storage.Root, w.cfg.Storage.HLSDir, tc.Slug)
	log.Debug().Str("input_path", inputPath).Str("output_dir", outputDir).Msg("paths resolved")

	if _, err := os.Stat(inputPath); err != nil {
		log.Error().Err(err).Str("input_path", inputPath).Msg("input file not accessible")
		w.publishFailEvent(mediaSlug, mediaUserID, "input file not accessible")
		w.failTranscodeAndJob(ctx, tc.ID, job.ID, "input file not accessible")
		return
	}

	if err := os.MkdirAll(outputDir, 0o755); err != nil {
		log.Error().Err(err).Str("output_dir", outputDir).Msg("create output dir")
		w.publishFailEvent(mediaSlug, mediaUserID, "create output dir failed")
		w.failTranscodeAndJob(ctx, tc.ID, job.ID, "create output dir failed")
		return
	}

	// Probe duration
	log.Debug().Msg("probing video duration")
	duration, err := ProbeDuration(w.cfg.FFmpeg.FFprobePath, inputPath)
	if err != nil {
		log.Warn().Err(err).Int32("duration_sec", duration).Msg("probe duration failed, continuing without duration")
	} else {
		log.Debug().Int32("duration_sec", duration).Msg("duration probed")
	}

	// Build and run FFmpeg
	args := BuildFFmpegArgs(inputPath, outputDir)
	log.Debug().Strs("ffmpeg_args", args).Msg("running ffmpeg")

	cmd := exec.CommandContext(ctx, w.cfg.FFmpeg.Path, args...)

	stderr, err := cmd.StderrPipe()
	if err != nil {
		log.Error().Err(err).Msg("create stderr pipe")
		w.publishFailEvent(mediaSlug, mediaUserID, "stderr pipe failed")
		w.failTranscodeAndJob(ctx, tc.ID, job.ID, "stderr pipe failed")
		return
	}

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		log.Error().Err(err).Msg("create stdout pipe")
		w.publishFailEvent(mediaSlug, mediaUserID, "stdout pipe failed")
		w.failTranscodeAndJob(ctx, tc.ID, job.ID, "stdout pipe failed")
		return
	}

	if err := cmd.Start(); err != nil {
		log.Error().Err(err).Msg("start ffmpeg")
		w.publishFailEvent(mediaSlug, mediaUserID, "ffmpeg start failed")
		w.failTranscodeAndJob(ctx, tc.ID, job.ID, "ffmpeg start failed")
		return
	}
	log.Debug().Int("pid", cmd.Process.Pid).Msg("ffmpeg started")

	// Capture stderr to log on failure
	go func() {
		stderrLog := w.logger.With().Str("job_slug", job.Slug).Str("stream", "stderr").Logger()
		scanner := bufio.NewScanner(stderr)
		for scanner.Scan() {
			stderrLog.Debug().Str("line", scanner.Text()).Msg("ffmpeg")
		}
	}()

	// Parse progress
	scanner := bufio.NewScanner(stdout)
	nextLogPct := 0
	ParseProgress(scanner, duration, func(pct int) {
		_ = w.cache.MediaProgress.SetProgress(ctx, tc.Slug, pct)
		if pct >= nextLogPct {
			log.Debug().Int("progress_pct", pct).Msg("transcode progress")
			nextLogPct = pct + 25
		}
		if w.broadcaster != nil && mediaSlug != "" {
			w.broadcaster.Publish(ProgressEvent{
				UserID:    mediaUserID,
				MediaSlug: mediaSlug,
				Progress:  pct,
				Status:    "processing",
			})
		}
	})

	if err := cmd.Wait(); err != nil {
		log.Error().Err(err).Msg("ffmpeg exited with error")
		_ = os.RemoveAll(outputDir)
		w.publishFailEvent(mediaSlug, mediaUserID, fmt.Sprintf("ffmpeg failed: %v", err))
		w.failTranscodeAndJob(ctx, tc.ID, job.ID, fmt.Sprintf("ffmpeg failed: %v", err))
		return
	}
	log.Debug().Msg("ffmpeg completed successfully")

	// Extract poster image
	posterPath := filepath.Join(outputDir, "poster.jpg")
	if err := ExtractPoster(w.cfg.FFmpeg.Path, inputPath, posterPath, duration); err != nil {
		log.Warn().Err(err).Msg("extract poster failed, continuing without poster")
		posterPath = ""
	}

	// Update media item poster path
	if posterPath != "" && miErr == nil {
		_ = w.queries.UpdateMediaItemPoster(ctx, sqlc.UpdateMediaItemPosterParams{
			ID:         mi.ID,
			PosterPath: pgtype.Text{String: posterPath, Valid: true},
		})
	}

	// Success
	if err := w.queries.UpdateTranscodeHLS(ctx, sqlc.UpdateTranscodeHLSParams{
		ID:          tc.ID,
		HlsDir:      pgtype.Text{String: outputDir, Valid: true},
		DurationSec: pgtype.Int4{Int32: duration, Valid: duration > 0},
	}); err != nil {
		log.Error().Err(err).Msg("update transcode HLS")
		return
	}

	if err := w.queries.UpdateTranscodeStatus(ctx, sqlc.UpdateTranscodeStatusParams{
		ID:     tc.ID,
		Status: "done",
	}); err != nil {
		log.Error().Err(err).Msg("update transcode status")
		return
	}

	if err := w.queries.UpdateJobStatus(ctx, sqlc.UpdateJobStatusParams{
		ID:     job.ID,
		Status: "done",
	}); err != nil {
		log.Error().Err(err).Msg("update job status")
		return
	}

	_ = w.cache.MediaProgress.DeleteProgress(ctx, tc.Slug)
	log.Info().Int32("duration", duration).Msg("transcode completed")

	if w.broadcaster != nil && mediaSlug != "" {
		w.broadcaster.Publish(ProgressEvent{
			UserID:    mediaUserID,
			MediaSlug: mediaSlug,
			Progress:  100,
			Status:    "done",
		})
	}
}

func (w *Worker) publishFailEvent(mediaSlug string, userID int64, errMsg string) {
	if w.broadcaster == nil || mediaSlug == "" {
		return
	}
	w.broadcaster.Publish(ProgressEvent{
		UserID:    userID,
		MediaSlug: mediaSlug,
		Status:    "failed",
		ErrorMsg:  errMsg,
	})
}

func (w *Worker) failJob(ctx context.Context, jobID int64, msg string) {
	_ = w.queries.UpdateJobStatus(ctx, sqlc.UpdateJobStatusParams{
		ID:       jobID,
		Status:   "failed",
		ErrorMsg: pgtype.Text{String: msg, Valid: true},
	})
}

func (w *Worker) failTranscodeAndJob(ctx context.Context, transcodeID, jobID int64, msg string) {
	_ = w.queries.UpdateTranscodeStatus(ctx, sqlc.UpdateTranscodeStatusParams{
		ID:       transcodeID,
		Status:   "failed",
		ErrorMsg: pgtype.Text{String: msg, Valid: true},
	})
	w.failJob(ctx, jobID, msg)
}
