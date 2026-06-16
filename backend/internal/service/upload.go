package service

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"math"
	"math/rand"
	"time"

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

type UploadService struct {
	queries *sqlc.Queries
	pg      *pgxpool.Pool
	cfg     *config.Config
	store   *storage.Local
	cache   *cache.Cache
	logger  zerolog.Logger
}

func NewUploadService(queries *sqlc.Queries, pg *pgxpool.Pool, cfg *config.Config, store *storage.Local, c *cache.Cache, logger zerolog.Logger) *UploadService {
	return &UploadService{queries: queries, pg: pg, cfg: cfg, store: store, cache: c, logger: logger}
}

type PreCheckRequest struct {
	PreHash  string `json:"preHash"`
	FileSize int64  `json:"fileSize"`
}

type PreCheckResponse struct {
	Status string `json:"status"`
}

type RequestChallengeRequest struct {
	FileHash string `json:"fileHash"`
}

type RequestChallengeResponse struct {
	Status          string `json:"status"`
	ChallengeOffset int64  `json:"challengeOffset,omitempty"`
	ChallengeToken  string `json:"challengeToken,omitempty"`
}

type VerifyRequest struct {
	FileHash  string `json:"fileHash"`
	ProofCode string `json:"proofCode"`
}

type ExistingFileRef struct {
	FileName string `json:"fileName"`
	Path     string `json:"path"`
}

type VerifyResponse struct {
	Status           string            `json:"status"`
	PhysicalFileSlug string            `json:"physicalFileSlug,omitempty"`
	ExistingFiles    []ExistingFileRef `json:"existingFiles,omitempty"`
}

type InitRequest struct {
	FileHash   string `json:"fileHash"`
	PreHash    string `json:"preHash"`
	FileSize   int64  `json:"fileSize"`
	MimeType   string `json:"mimeType"`
	FileName   string `json:"fileName"`
	ParentSlug string `json:"parentSlug"`
}

type InitResponse struct {
	UploadSlug      string `json:"uploadSlug"`
	TotalChunks     int32  `json:"totalChunks"`
	ChunkSize       int32  `json:"chunkSize"`
	CompletedChunks []int  `json:"completedChunks"`
}

type CompleteResponse struct {
	Status           string `json:"status"`
	PhysicalFileSlug string `json:"physicalFileSlug,omitempty"`
}

type StatusResponse struct {
	Status           string  `json:"status"`
	PhysicalFileSlug string  `json:"physicalFileSlug,omitempty"`
	Error            *string `json:"error,omitempty"`
}

func (s *UploadService) PreCheck(ctx context.Context, userID int64, req PreCheckRequest) (*PreCheckResponse, error) {
	if req.FileSize <= 0 {
		return nil, model.ErrInvalidInput
	}
	if req.FileSize > s.cfg.Storage.MaxUploadSize {
		return nil, model.ErrFileTooLarge
	}

	// Check Redis pre-cache
	slug, err := s.cache.PreCache.Get(ctx, req.FileSize, req.PreHash)
	if err == nil {
		s.logger.Debug().Int64("userID", userID).Str("preHash", req.PreHash[:8]+"...").Int64("fileSize", req.FileSize).Str("cachedSlug", slug).Msg("pre-check cache hit")
		return &PreCheckResponse{Status: "SUSPECT_HIT"}, nil
	}

	// Check DB
	pf, err := s.queries.GetPhysicalFileByPreHash(ctx, sqlc.GetPhysicalFileByPreHashParams{
		PreHash:  req.PreHash,
		FileSize: req.FileSize,
	})
	if err == nil {
		s.logger.Debug().Int64("userID", userID).Str("preHash", req.PreHash[:8]+"...").Int64("fileSize", req.FileSize).Int64("physicalID", pf.ID).Msg("pre-check db hit")
		_ = s.cache.PreCache.Set(ctx, req.FileSize, req.PreHash, "")
		return &PreCheckResponse{Status: "SUSPECT_HIT"}, nil
	}

	s.logger.Debug().Int64("userID", userID).Str("preHash", req.PreHash[:8]+"...").Int64("fileSize", req.FileSize).Msg("pre-check miss")
	return &PreCheckResponse{Status: "NOT_FOUND"}, nil
}

func (s *UploadService) RequestChallenge(ctx context.Context, userID int64, req RequestChallengeRequest) (*RequestChallengeResponse, error) {
	if req.FileHash == "" {
		return nil, model.ErrInvalidInput
	}

	pf, err := s.queries.GetPhysicalFileByHash(ctx, sqlc.GetPhysicalFileByHashParams{
		HashAlgo: "sha256",
		FileHash: req.FileHash,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			s.logger.Warn().Int64("userID", userID).Str("fileHash", req.FileHash[:8]+"...").Msg("request challenge: file not found")
			return &RequestChallengeResponse{Status: "NOT_FOUND"}, nil
		}
		return nil, fmt.Errorf("get physical file: %w", err)
	}

	// If a challenge already exists for this user+hash, reuse it so concurrent
	// uploads of the same file share one challenge instead of overwriting each other.
	existingOffset, existingToken, err := s.cache.Challenge.GetChallenge(ctx, userID, req.FileHash)
	if err == nil {
		s.logger.Debug().Int64("userID", userID).Str("fileHash", req.FileHash[:8]+"...").Int("offset", existingOffset).Msg("request challenge: reusing existing challenge")
		return &RequestChallengeResponse{
			Status:          "CHALLENGE",
			ChallengeOffset: int64(existingOffset),
			ChallengeToken:  existingToken,
		}, nil
	}

	offset := int(rand.Int63n(max(1, pf.FileSize-1024)))
	token, err := gonanoid.New(32)
	if err != nil {
		return nil, fmt.Errorf("generate token: %w", err)
	}

	if err := s.cache.Challenge.SetChallenge(ctx, userID, req.FileHash, offset, token); err != nil {
		s.logger.Error().Int64("userID", userID).Str("fileHash", req.FileHash[:8]+"...").Err(err).Msg("request challenge: set challenge failed")
		return nil, fmt.Errorf("set challenge: %w", err)
	}

	s.logger.Info().Int64("userID", userID).Str("fileHash", req.FileHash[:8]+"...").Int("offset", offset).Int64("fileSize", pf.FileSize).Msg("request challenge: issued")
	return &RequestChallengeResponse{
		Status:          "CHALLENGE",
		ChallengeOffset: int64(offset),
		ChallengeToken:  token,
	}, nil
}

func (s *UploadService) Verify(ctx context.Context, userID int64, req VerifyRequest) (*VerifyResponse, error) {
	if req.FileHash == "" || req.ProofCode == "" {
		return nil, model.ErrInvalidInput
	}

	offset, token, err := s.cache.Challenge.ConsumeChallenge(ctx, userID, req.FileHash)
	if err != nil {
		s.logger.Warn().Int64("userID", userID).Str("fileHash", req.FileHash[:8]+"...").Err(err).Msg("verify: consume challenge failed, treating as MISS")
		return &VerifyResponse{Status: "MISS"}, nil
	}

	diskBytes, err := s.store.ReadAt(req.FileHash, int64(offset), 1024)
	if err != nil {
		s.logger.Error().Str("fileHash", req.FileHash[:8]+"...").Int("offset", offset).Err(err).Msg("verify: read file for challenge failed")
		return nil, fmt.Errorf("read file: %w", err)
	}

	h := sha256.New()
	h.Write(diskBytes)
	h.Write([]byte(token))
	expected := hex.EncodeToString(h.Sum(nil))

	if expected != req.ProofCode {
		s.logger.Warn().Int64("userID", userID).Str("fileHash", req.FileHash[:8]+"...").Int("offset", offset).Msg("verify: proof mismatch")
		return &VerifyResponse{Status: "MISS"}, nil
	}

	pf, err := s.queries.GetPhysicalFileByHash(ctx, sqlc.GetPhysicalFileByHashParams{
		HashAlgo: "sha256",
		FileHash: req.FileHash,
	})
	if err != nil {
		s.logger.Error().Str("fileHash", req.FileHash[:8]+"...").Err(err).Msg("verify: get physical file after proof match failed")
		return nil, fmt.Errorf("get physical file: %w", err)
	}

	// Find existing user files referencing this physical file
	existingFiles := make([]ExistingFileRef, 0)
	rows, err := s.queries.GetUserFilesByPhysicalFileID(ctx, sqlc.GetUserFilesByPhysicalFileIDParams{
		PhysicalFileID: pgtype.Int8{Int64: pf.ID, Valid: true},
		UserID:         userID,
	})
	if err == nil {
		for _, r := range rows {
			path := r.FileName
			if r.ParentSlug.Valid && r.ParentSlug.String != "" {
				path = r.ParentSlug.String + "/" + r.FileName
			}
			existingFiles = append(existingFiles, ExistingFileRef{
				FileName: r.FileName,
				Path:     path,
			})
		}
	}

	s.logger.Info().Int64("userID", userID).Str("fileHash", req.FileHash[:8]+"...").Int("existingFiles", len(existingFiles)).Msg("verify: hit")
	return &VerifyResponse{
		Status:           "HIT",
		PhysicalFileSlug: pf.Slug,
		ExistingFiles:    existingFiles,
	}, nil
}

func (s *UploadService) Init(ctx context.Context, userID int64, sessionID string, req InitRequest) (*InitResponse, error) {
	if req.PreHash == "" || req.FileSize <= 0 || req.MimeType == "" {
		s.logger.Warn().Int64("userID", userID).Str("preHash", req.PreHash).Int64("fileSize", req.FileSize).Str("mimeType", req.MimeType).Str("fileName", req.FileName).Msg("init: invalid input - missing required fields")
		return nil, model.ErrInvalidInput
	}
	if req.FileSize > s.cfg.Storage.MaxUploadSize {
		return nil, model.ErrFileTooLarge
	}
	if err := s.ensureUploadParentUnlocked(ctx, userID, sessionID, req.ParentSlug); err != nil {
		return nil, err
	}

	// Resume check only when hash is known
	if req.FileHash != "" {
		existing, err := s.queries.GetUploadTaskByHashForUser(ctx, sqlc.GetUploadTaskByHashForUserParams{
			OwnerUserID: userID,
			FileHash:    req.FileHash,
		})
		if err != nil {
			if !errors.Is(err, pgx.ErrNoRows) {
				s.logger.Warn().Int64("userID", userID).Str("fileHash", req.FileHash[:8]+"...").Err(err).Msg("init: resume query failed")
			}
		} else if existing.ID != 0 && existing.Status != "done" && existing.Status != "failed" {
			s.logger.Info().Str("slug", existing.Slug).Int64("existingTaskID", existing.ID).Str("status", existing.Status).Int32("totalChunks", existing.TotalChunks).Int32("chunkSize", existing.ChunkSize).Int64("fileSize", existing.FileSize).Msg("init: found existing upload task, trying resume")
			chunks, chunkErr := s.cache.Chunks.ListChunks(ctx, existing.Slug)
			if chunkErr != nil {
				s.logger.Warn().Str("slug", existing.Slug).Err(chunkErr).Msg("init: list cached chunks failed")
			}
			validChunks, err := s.store.ValidChunkSet(existing.Slug, int(existing.TotalChunks), int64(existing.ChunkSize), existing.FileSize)
			if err != nil {
				s.logger.Warn().Str("slug", existing.Slug).Err(err).Msg("init: validate resume chunks failed")
				validChunks = map[int]bool{}
			}
			filteredChunks := make([]int, 0, len(chunks))
			seen := make(map[int]bool, len(chunks))
			for _, chunk := range chunks {
				if validChunks[chunk] && !seen[chunk] {
					filteredChunks = append(filteredChunks, chunk)
					seen[chunk] = true
				}
			}
			s.logger.Info().Str("slug", existing.Slug).Str("status", existing.Status).Int("cachedChunks", len(chunks)).Int("validChunks", len(filteredChunks)).Int("totalChunks", int(existing.TotalChunks)).Msg("init: resuming existing upload")
			return &InitResponse{
				UploadSlug:      existing.Slug,
				TotalChunks:     existing.TotalChunks,
				ChunkSize:       existing.ChunkSize,
				CompletedChunks: filteredChunks,
			}, nil
		} else if existing.ID != 0 {
			s.logger.Debug().Int64("userID", userID).Str("fileHash", req.FileHash[:8]+"...").Str("status", existing.Status).Int64("existingTaskID", existing.ID).Msg("init: existing task found but not resumable (done/failed)")
		}
	}

	totalChunks := int32(math.Ceil(float64(req.FileSize) / float64(s.cfg.Upload.ChunkSize)))

	slug, err := gonanoid.New(21)
	if err != nil {
		return nil, fmt.Errorf("generate slug: %w", err)
	}

	_, err = s.queries.CreateUploadTask(ctx, sqlc.CreateUploadTaskParams{
		Slug:         slug,
		OwnerUserID:  userID,
		HashAlgo:     "sha256",
		FileHash:     req.FileHash,
		PreHash:      req.PreHash,
		FileSize:     req.FileSize,
		MimeType:     req.MimeType,
		OriginalName: req.FileName,
		TotalChunks:  totalChunks,
		ChunkSize:    s.cfg.Upload.ChunkSize,
		Status:       "uploading",
		ExpiresAt:    pgtype.Timestamptz{Time: time.Now().Add(time.Duration(s.cfg.Upload.TaskExpiryDays) * 24 * time.Hour), Valid: true},
		ParentSlug:   pgtype.Text{String: req.ParentSlug, Valid: req.ParentSlug != ""},
	})
	if err != nil {
		return nil, fmt.Errorf("create upload task: %w", err)
	}

	s.logger.Info().Str("slug", slug).Int64("userID", userID).Str("fileName", req.FileName).Int64("fileSize", req.FileSize).Int32("totalChunks", totalChunks).Msg("init: new upload task created")
	return &InitResponse{
		UploadSlug:      slug,
		TotalChunks:     totalChunks,
		ChunkSize:       s.cfg.Upload.ChunkSize,
		CompletedChunks: []int{},
	}, nil
}

func (s *UploadService) AppendChunk(ctx context.Context, userID int64, uploadSlug string, chunkIndex int32, data []byte) error {
	s.logger.Debug().Int64("userID", userID).Str("slug", uploadSlug).Int32("chunkIndex", chunkIndex).Int("dataSize", len(data)).Msg("append chunk: entry")
	task, err := s.queries.GetUploadTaskBySlug(ctx, uploadSlug)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			s.logger.Warn().Int64("userID", userID).Str("slug", uploadSlug).Int32("chunkIndex", chunkIndex).Msg("append chunk: task not found")
			return model.ErrNotFound
		}
		s.logger.Error().Int64("userID", userID).Str("slug", uploadSlug).Int32("chunkIndex", chunkIndex).Err(err).Msg("append chunk: get task failed")
		return fmt.Errorf("get task: %w", err)
	}

	if task.OwnerUserID != userID {
		s.logger.Warn().Int64("userID", userID).Int64("taskOwnerID", task.OwnerUserID).Str("slug", uploadSlug).Int32("chunkIndex", chunkIndex).Msg("append chunk: unauthorized - owner mismatch")
		return model.ErrUnauthorized
	}
	if chunkIndex < 0 || chunkIndex >= task.TotalChunks {
		s.logger.Warn().Int64("userID", userID).Str("slug", uploadSlug).Int32("chunkIndex", chunkIndex).Int32("totalChunks", task.TotalChunks).Msg("append chunk: invalid chunk index")
		return model.ErrInvalidInput
	}

	isLast := chunkIndex == task.TotalChunks-1
	expectedSize := int(task.ChunkSize)
	if isLast {
		remaining := int(task.FileSize) - int(chunkIndex)*int(task.ChunkSize)
		expectedSize = remaining
	}
	if len(data) != expectedSize {
		s.logger.Warn().Int64("userID", userID).Str("slug", uploadSlug).Int32("chunkIndex", chunkIndex).Int("expectedSize", expectedSize).Int("gotSize", len(data)).Bool("isLast", isLast).Int32("taskChunkSize", task.ChunkSize).Int64("taskFileSize", task.FileSize).Int32("taskTotalChunks", task.TotalChunks).Msg("append chunk: size mismatch")
		return model.ErrInvalidInput
	}

	// Tolerate chunks arriving after the task has moved to merging/done:
	// if the chunk already exists with correct size, skip the write and
	// just ensure it's tracked in the chunk cache.
	if task.Status != "uploading" {
		s.logger.Warn().Int64("userID", userID).Str("slug", uploadSlug).Int32("chunkIndex", chunkIndex).Str("taskStatus", task.Status).Int32("taskTotalChunks", task.TotalChunks).Msg("append chunk: task not in uploading status, checking if chunk already exists")
		if s.store.ChunkExists(uploadSlug, int(chunkIndex), int64(expectedSize)) {
			s.logger.Debug().Str("slug", uploadSlug).Int32("chunkIndex", chunkIndex).Str("status", task.Status).Msg("append chunk: task no longer uploading, chunk already exists, skipping")
			_ = s.cache.Chunks.AddChunk(ctx, uploadSlug, int(chunkIndex))
			return nil
		}
		s.logger.Warn().Int64("userID", userID).Str("slug", uploadSlug).Int32("chunkIndex", chunkIndex).Str("taskStatus", task.Status).Int("expectedSize", expectedSize).Int64("taskID", task.ID).Msg("append chunk: chunk missing and task no longer uploading - rejecting")
		return model.ErrInvalidInput
	}

	s.logger.Debug().Int64("userID", userID).Str("slug", uploadSlug).Int32("chunkIndex", chunkIndex).Int("dataSize", len(data)).Int64("taskID", task.ID).Msg("append chunk: writing chunk")
	if err := s.store.WriteChunk(uploadSlug, int(chunkIndex), bytes.NewReader(data)); err != nil {
		s.logger.Error().Int64("userID", userID).Str("slug", uploadSlug).Int32("chunkIndex", chunkIndex).Int("dataSize", len(data)).Int64("taskID", task.ID).Err(err).Msg("append chunk: write chunk failed")
		return fmt.Errorf("write chunk: %w", err)
	}

	if err := s.cache.Chunks.AddChunk(ctx, uploadSlug, int(chunkIndex)); err != nil {
		s.logger.Error().Int64("userID", userID).Str("slug", uploadSlug).Int32("chunkIndex", chunkIndex).Err(err).Msg("append chunk: track chunk in cache failed")
		return fmt.Errorf("track chunk: %w", err)
	}

	s.logger.Debug().Int64("userID", userID).Str("slug", uploadSlug).Int32("chunkIndex", chunkIndex).Int("size", len(data)).Int32("totalChunks", task.TotalChunks).Msg("append chunk: success")
	return nil
}

type UpdateHashRequest struct {
	UploadSlug string `json:"uploadSlug"`
	FileHash   string `json:"fileHash"`
	PreHash    string `json:"preHash"`
}

func (s *UploadService) UpdateHash(ctx context.Context, userID int64, req UpdateHashRequest) error {
	if req.FileHash == "" || req.UploadSlug == "" {
		return model.ErrInvalidInput
	}
	task, err := s.queries.GetUploadTaskBySlug(ctx, req.UploadSlug)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return model.ErrNotFound
		}
		return fmt.Errorf("get task: %w", err)
	}
	if task.OwnerUserID != userID {
		return model.ErrUnauthorized
	}
	if task.FileHash != "" {
		s.logger.Debug().Str("slug", req.UploadSlug).Str("existingHash", task.FileHash[:8]+"...").Msg("update-hash: already set, skipping")
		return nil // already set
	}
	if err := s.queries.UpdateUploadTaskFileHash(ctx, sqlc.UpdateUploadTaskFileHashParams{
		ID:       task.ID,
		FileHash: req.FileHash,
	}); err != nil {
		s.logger.Error().Str("slug", req.UploadSlug).Err(err).Msg("update-hash: db update failed")
		return fmt.Errorf("update hash: %w", err)
	}
	if req.PreHash != "" {
		_ = s.queries.UpdateUploadTaskPreHash(ctx, sqlc.UpdateUploadTaskPreHashParams{
			ID:      task.ID,
			PreHash: req.PreHash,
		})
	}
	s.logger.Info().Str("slug", req.UploadSlug).Str("fileHash", req.FileHash[:8]+"...").Msg("update-hash: done")
	return nil
}

func (s *UploadService) Complete(ctx context.Context, userID int64, uploadSlug string) (*CompleteResponse, error) {
	s.logger.Info().Int64("userID", userID).Str("slug", uploadSlug).Msg("complete: entry")
	task, err := s.queries.GetUploadTaskBySlug(ctx, uploadSlug)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			s.logger.Warn().Int64("userID", userID).Str("slug", uploadSlug).Msg("complete: task not found")
			return nil, model.ErrNotFound
		}
		s.logger.Error().Int64("userID", userID).Str("slug", uploadSlug).Err(err).Msg("complete: get task failed")
		return nil, fmt.Errorf("get task: %w", err)
	}

	if task.OwnerUserID != userID {
		s.logger.Warn().Int64("userID", userID).Int64("taskOwnerID", task.OwnerUserID).Str("slug", uploadSlug).Msg("complete: unauthorized - owner mismatch")
		return nil, model.ErrUnauthorized
	}

	s.logger.Info().Int64("userID", userID).Str("slug", uploadSlug).Int64("taskID", task.ID).Str("status", task.Status).Int32("totalChunks", task.TotalChunks).Int32("chunkSize", task.ChunkSize).Int64("fileSize", task.FileSize).Str("fileHash", safeHashPrefix(task.FileHash)).Str("preHash", safeHashPrefix(task.PreHash)).Msg("complete: task loaded")

	if task.Status != "uploading" {
		s.logger.Warn().Int64("userID", userID).Str("slug", uploadSlug).Str("status", task.Status).Int64("taskID", task.ID).Msg("complete: called on non-uploading task")
		if task.Status == "done" {
			pf, pfErr := s.queries.GetPhysicalFileByID(ctx, task.PhysicalFileID.Int64)
			if pfErr == nil {
				s.logger.Info().Int64("userID", userID).Str("slug", uploadSlug).Str("physicalFileSlug", pf.Slug).Msg("complete: already done, returning existing result")
				return &CompleteResponse{Status: "DONE", PhysicalFileSlug: pf.Slug}, nil
			}
			s.logger.Warn().Int64("userID", userID).Str("slug", uploadSlug).Int64("physicalFileID", task.PhysicalFileID.Int64).Err(pfErr).Msg("complete: status is done but could not load physical file")
		}
		return &CompleteResponse{Status: task.Status}, nil
	}

	s.logger.Info().Int64("userID", userID).Str("slug", uploadSlug).Int32("totalChunks", task.TotalChunks).Msg("complete: validating chunks")
	chunkIssues, err := s.store.ValidateChunks(uploadSlug, int(task.TotalChunks), int64(task.ChunkSize), task.FileSize)
	if err != nil {
		s.logger.Error().Int64("userID", userID).Str("slug", uploadSlug).Int32("totalChunks", task.TotalChunks).Err(err).Msg("complete: validate chunks failed")
		return nil, fmt.Errorf("validate chunks: %w", err)
	}
	if len(chunkIssues) > 0 {
		issue := chunkIssues[0]
		issueEvent := s.logger.Warn().Int64("userID", userID).Str("slug", uploadSlug).Int("issueCount", len(chunkIssues)).Int("firstChunkIndex", issue.Index).Int64("expected", issue.Expected).Bool("missing", issue.Missing)
		if !issue.Missing {
			issueEvent = issueEvent.Int64("actual", issue.Actual)
		}
		// Log all issues in detail for debugging
		for idx, iss := range chunkIssues {
			if idx >= 10 {
				issueEvent = issueEvent.Int("moreIssues", len(chunkIssues)-10)
				break
			}
			issueEvent = issueEvent.Int(fmt.Sprintf("issue_%d_index", idx), iss.Index)
			if iss.Missing {
				issueEvent = issueEvent.Bool(fmt.Sprintf("issue_%d_missing", idx), true)
			}
			issueEvent = issueEvent.Int64(fmt.Sprintf("issue_%d_expected", idx), iss.Expected)
			if !iss.Missing {
				issueEvent = issueEvent.Int64(fmt.Sprintf("issue_%d_actual", idx), iss.Actual)
			}
		}
		issueEvent.Msg("complete: chunks validation failed - some chunks are missing or size mismatch")
		return nil, model.ErrInvalidInput
	}
	s.logger.Info().Int64("userID", userID).Str("slug", uploadSlug).Int32("totalChunks", task.TotalChunks).Msg("complete: all chunks validated OK")

	// Atomic status transition: only proceed if still "uploading"
	s.logger.Debug().Int64("userID", userID).Str("slug", uploadSlug).Int64("taskID", task.ID).Msg("complete: executing atomic status update uploading->merging")
	tag, err := s.pg.Exec(ctx,
		`UPDATE upload_tasks SET status = 'merging' WHERE id = $1 AND status = 'uploading'`,
		task.ID,
	)
	if err != nil {
		s.logger.Error().Int64("userID", userID).Str("slug", uploadSlug).Int64("taskID", task.ID).Err(err).Msg("complete: atomic status update failed")
		return nil, fmt.Errorf("atomic status update: %w", err)
	}
	if tag.RowsAffected() == 0 {
		s.logger.Warn().Int64("userID", userID).Str("slug", uploadSlug).Int64("taskID", task.ID).Msg("complete: status already changed by another request (race)")
		return &CompleteResponse{Status: "MERGING"}, nil
	}
	s.logger.Info().Int64("userID", userID).Str("slug", uploadSlug).Int64("taskID", task.ID).Msg("complete: status transitioned to merging")

	fileHash := task.FileHash
	var storagePath string

	if fileHash == "" {
		s.logger.Info().Int64("userID", userID).Str("slug", uploadSlug).Int64("taskID", task.ID).Int32("totalChunks", task.TotalChunks).Msg("complete: hash not known, merging chunks and computing hash")
		computedHash, sp, err := s.store.MergeChunksAndHash(uploadSlug, int(task.TotalChunks))
		if err != nil {
			_ = s.queries.UpdateUploadTaskStatus(ctx, sqlc.UpdateUploadTaskStatusParams{ID: task.ID, Status: "failed"})
			s.logger.Error().Int64("userID", userID).Str("slug", uploadSlug).Int64("taskID", task.ID).Int32("totalChunks", task.TotalChunks).Err(err).Msg("complete: merge and hash failed")
			return nil, fmt.Errorf("merge and hash: %w", err)
		}
		fileHash = computedHash
		storagePath = sp
		_ = s.queries.UpdateUploadTaskFileHash(ctx, sqlc.UpdateUploadTaskFileHashParams{ID: task.ID, FileHash: fileHash})
		s.logger.Info().Int64("userID", userID).Str("slug", uploadSlug).Str("fileHash", fileHash[:16]+"...").Str("storagePath", storagePath).Msg("complete: hash computed from merged chunks")
	} else {
		s.logger.Info().Int64("userID", userID).Str("slug", uploadSlug).Str("fileHash", fileHash[:16]+"...").Int32("totalChunks", task.TotalChunks).Msg("complete: merging known hash")
		acquired, err := s.cache.Lock.AcquireMergeLock(ctx, fileHash)
		if err != nil {
			s.logger.Error().Int64("userID", userID).Str("slug", uploadSlug).Str("fileHash", fileHash[:16]+"...").Err(err).Msg("complete: acquire merge lock error")
			return nil, fmt.Errorf("acquire lock: %w", err)
		}
		if !acquired {
			s.logger.Info().Int64("userID", userID).Str("slug", uploadSlug).Str("fileHash", fileHash[:16]+"...").Msg("complete: merge lock not acquired, another instance merging")
			return &CompleteResponse{Status: "MERGING"}, nil
		}
		s.logger.Debug().Int64("userID", userID).Str("slug", uploadSlug).Str("fileHash", fileHash[:16]+"...").Msg("complete: merge lock acquired")
		defer s.cache.Lock.ReleaseMergeLock(ctx, fileHash)

		sp, err := s.store.MergeChunks(uploadSlug, fileHash, int(task.TotalChunks))
		if err != nil {
			_ = s.queries.UpdateUploadTaskStatus(ctx, sqlc.UpdateUploadTaskStatusParams{ID: task.ID, Status: "failed"})
			s.logger.Error().Int64("userID", userID).Str("slug", uploadSlug).Str("fileHash", fileHash[:16]+"...").Int64("taskID", task.ID).Err(err).Msg("complete: merge chunks failed")
			return nil, fmt.Errorf("merge chunks: %w", err)
		}
		storagePath = sp
		s.logger.Debug().Int64("userID", userID).Str("slug", uploadSlug).Str("fileHash", fileHash[:16]+"...").Str("storagePath", storagePath).Msg("complete: chunks merged successfully")
	}

	pf, err := s.queries.CreatePhysicalFile(ctx, sqlc.CreatePhysicalFileParams{
		Slug:        task.Slug,
		HashAlgo:    "sha256",
		FileHash:    fileHash,
		PreHash:     task.PreHash,
		FileSize:    task.FileSize,
		MimeType:    task.MimeType,
		StoragePath: storagePath,
		Status:      "completed",
	})
	if err != nil {
		s.logger.Warn().Str("slug", uploadSlug).Str("fileHash", fileHash[:16]+"...").Err(err).Msg("create physical file failed (race?), trying to get existing")
		pf, err = s.queries.GetPhysicalFileByHash(ctx, sqlc.GetPhysicalFileByHashParams{
			HashAlgo: "sha256",
			FileHash: fileHash,
		})
		if err != nil {
			s.logger.Error().Str("slug", uploadSlug).Err(err).Msg("get physical file after create conflict also failed")
			return nil, fmt.Errorf("get physical file: %w", err)
		}
	}

	if err := s.queries.UpdateUploadTaskPhysicalFile(ctx, sqlc.UpdateUploadTaskPhysicalFileParams{
		ID:             task.ID,
		PhysicalFileID: pgtype.Int8{Int64: pf.ID, Valid: true},
	}); err != nil {
		return nil, fmt.Errorf("link physical file to task: %w", err)
	}
	if err := s.queries.UpdateUploadTaskStatus(ctx, sqlc.UpdateUploadTaskStatusParams{
		ID:     task.ID,
		Status: "done",
	}); err != nil {
		return nil, fmt.Errorf("update task status: %w", err)
	}

	_ = s.cache.PreCache.Set(ctx, task.FileSize, task.PreHash, pf.Slug)
	_ = s.store.CleanupUpload(uploadSlug)
	_ = s.cache.Chunks.DeleteChunks(ctx, uploadSlug)

	s.logger.Info().Int64("userID", userID).Str("slug", uploadSlug).Str("physicalFileSlug", pf.Slug).Int64("fileSize", task.FileSize).Int64("pfID", pf.ID).Str("fileHash", safeHashPrefix(fileHash)).Msg("complete: done")
	return &CompleteResponse{
		Status:           "DONE",
		PhysicalFileSlug: pf.Slug,
	}, nil
}

func safeHashPrefix(hash string) string {
	if len(hash) >= 16 {
		return hash[:16] + "..."
	}
	if len(hash) > 0 {
		return hash
	}
	return "(empty)"
}

func (s *UploadService) GetStatus(ctx context.Context, userID int64, uploadSlug string) (*StatusResponse, error) {
	task, err := s.queries.GetUploadTaskBySlug(ctx, uploadSlug)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, model.ErrNotFound
		}
		return nil, fmt.Errorf("get task: %w", err)
	}

	if task.OwnerUserID != userID {
		return nil, model.ErrUnauthorized
	}

	resp := &StatusResponse{Status: task.Status}

	if task.Status == "done" && task.PhysicalFileID.Valid {
		pf, err := s.queries.GetPhysicalFileByID(ctx, task.PhysicalFileID.Int64)
		if err == nil {
			resp.PhysicalFileSlug = pf.Slug
		}
	}

	if task.ErrorMsg.Valid {
		resp.Error = &task.ErrorMsg.String
	}

	s.logger.Debug().Str("slug", uploadSlug).Str("status", task.Status).Msg("get-status")
	return resp, nil
}

type TaskItem struct {
	Slug          string  `json:"slug"`
	FileName      string  `json:"fileName"`
	FileSize      int64   `json:"fileSize"`
	MimeType      string  `json:"mimeType"`
	Status        string  `json:"status"`
	ErrorMsg      string  `json:"errorMsg,omitempty"`
	TotalChunks   int32   `json:"totalChunks"`
	ReceivedBytes int64   `json:"receivedBytes"`
	ParentSlug    *string `json:"parentSlug,omitempty"`
	ParentName    *string `json:"parentName,omitempty"`
	CreatedAt     string  `json:"createdAt"`
	UpdatedAt     string  `json:"updatedAt"`
}

type ListTasksResponse struct {
	Items  []TaskItem `json:"items"`
	Total  int64      `json:"total"`
	Limit  int        `json:"limit"`
	Offset int        `json:"offset"`
}

func (s *UploadService) ListTasks(ctx context.Context, userID int64, limit, offset int, startDate, endDate pgtype.Timestamptz, status string) (*ListTasksResponse, error) {
	statusParam := pgtype.Text{Valid: false}
	if status != "" {
		statusParam = pgtype.Text{String: status, Valid: true}
	}

	total, err := s.queries.CountUploadTasksByUser(ctx, sqlc.CountUploadTasksByUserParams{
		OwnerUserID: userID,
		StartDate:   startDate,
		EndDate:     endDate,
		Status:      statusParam,
	})
	if err != nil {
		return nil, fmt.Errorf("count tasks: %w", err)
	}

	tasks, err := s.queries.ListUploadTasksByUser(ctx, sqlc.ListUploadTasksByUserParams{
		OwnerUserID: userID,
		Limit:       int32(limit),
		Offset:      int32(offset),
		StartDate:   startDate,
		EndDate:     endDate,
		Status:      statusParam,
	})
	if err != nil {
		return nil, fmt.Errorf("list tasks: %w", err)
	}

	items := make([]TaskItem, len(tasks))
	parentNameBySlug := make(map[string]*string)
	for i, t := range tasks {
		items[i] = TaskItem{
			Slug:        t.Slug,
			FileName:    t.OriginalName,
			FileSize:    t.FileSize,
			MimeType:    t.MimeType,
			Status:      t.Status,
			TotalChunks: t.TotalChunks,
			CreatedAt:   t.CreatedAt.Time.Format("2006-01-02T15:04:05Z"),
			UpdatedAt:   t.UpdatedAt.Time.Format("2006-01-02T15:04:05Z"),
		}
		if t.ErrorMsg.Valid {
			items[i].ErrorMsg = t.ErrorMsg.String
		}
		if t.ParentSlug.Valid {
			items[i].ParentSlug = &t.ParentSlug.String
			parentName, ok := parentNameBySlug[t.ParentSlug.String]
			if !ok {
				parentName = nil
				parent, err := s.queries.GetFileBySlugForUser(ctx, sqlc.GetFileBySlugForUserParams{
					Slug:   t.ParentSlug.String,
					UserID: userID,
				})
				if err == nil && parent.IsDir {
					name := parent.FileName
					parentName = &name
				} else if err != nil && !errors.Is(err, pgx.ErrNoRows) {
					s.logger.Warn().Err(err).Str("parentSlug", t.ParentSlug.String).Msg("resolve task parent name")
				}
				parentNameBySlug[t.ParentSlug.String] = parentName
			}
			items[i].ParentName = parentName
		}
		// Query chunk progress for interrupted tasks
		if t.Status == "uploading" || t.Status == "created" {
			chunkCount, err := s.cache.Chunks.ChunkCount(ctx, t.Slug)
			if err == nil && chunkCount > 0 {
				received := chunkCount * int64(t.ChunkSize)
				if received > t.FileSize {
					received = t.FileSize
				}
				items[i].ReceivedBytes = received
			}
		}
	}

	s.logger.Debug().Int64("userID", userID).Int64("total", total).Int("limit", limit).Int("offset", offset).Str("statusFilter", status).Msg("list-tasks")
	return &ListTasksResponse{
		Items:  items,
		Total:  total,
		Limit:  limit,
		Offset: offset,
	}, nil
}

func (s *UploadService) DeleteTask(ctx context.Context, userID int64, taskSlug string) error {
	task, err := s.queries.GetUploadTaskBySlug(ctx, taskSlug)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return model.ErrNotFound
		}
		return fmt.Errorf("get task: %w", err)
	}
	if task.OwnerUserID != userID {
		return model.ErrNotFound
	}

	_ = s.cache.Chunks.DeleteChunks(ctx, taskSlug)

	if err := s.queries.DeleteUploadTaskBySlug(ctx, sqlc.DeleteUploadTaskBySlugParams{
		Slug:        taskSlug,
		OwnerUserID: userID,
	}); err != nil {
		return fmt.Errorf("delete task: %w", err)
	}

	s.logger.Info().Str("slug", taskSlug).Str("status", task.Status).Msg("delete-task")
	return nil
}

func (s *UploadService) DeleteTasks(ctx context.Context, userID int64, slugs []string) error {
	if len(slugs) == 0 {
		return nil
	}

	for _, slug := range slugs {
		_ = s.cache.Chunks.DeleteChunks(ctx, slug)
	}

	if err := s.queries.DeleteUploadTasksBySlugs(ctx, sqlc.DeleteUploadTasksBySlugsParams{
		Column1:     slugs,
		OwnerUserID: userID,
	}); err != nil {
		return fmt.Errorf("delete tasks: %w", err)
	}

	s.logger.Info().Int64("userID", userID).Int("count", len(slugs)).Msg("delete-tasks")
	return nil
}

func (s *UploadService) RetryTask(ctx context.Context, userID int64, sessionID, taskSlug string) (*InitResponse, error) {
	task, err := s.queries.GetUploadTaskBySlug(ctx, taskSlug)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, model.ErrNotFound
		}
		return nil, fmt.Errorf("get task: %w", err)
	}
	if task.OwnerUserID != userID {
		return nil, model.ErrUnauthorized
	}
	if task.Status != "failed" {
		s.logger.Warn().Str("slug", taskSlug).Str("status", task.Status).Msg("retry-task: task not in failed status")
		return nil, model.ErrInvalidInput
	}
	if task.ParentSlug.Valid {
		if err := s.ensureUploadParentUnlocked(ctx, userID, sessionID, task.ParentSlug.String); err != nil {
			return nil, err
		}
	}

	totalChunks := int32(math.Ceil(float64(task.FileSize) / float64(s.cfg.Upload.ChunkSize)))

	slug, err := gonanoid.New(21)
	if err != nil {
		return nil, fmt.Errorf("generate slug: %w", err)
	}

	_, err = s.queries.CreateUploadTask(ctx, sqlc.CreateUploadTaskParams{
		Slug:         slug,
		OwnerUserID:  userID,
		HashAlgo:     "sha256",
		FileHash:     task.FileHash,
		PreHash:      task.PreHash,
		FileSize:     task.FileSize,
		MimeType:     task.MimeType,
		OriginalName: task.OriginalName,
		TotalChunks:  totalChunks,
		ChunkSize:    s.cfg.Upload.ChunkSize,
		Status:       "uploading",
		ExpiresAt:    pgtype.Timestamptz{Time: time.Now().Add(time.Duration(s.cfg.Upload.TaskExpiryDays) * 24 * time.Hour), Valid: true},
		ParentSlug:   task.ParentSlug,
	})
	if err != nil {
		return nil, fmt.Errorf("create upload task: %w", err)
	}

	s.logger.Info().Str("oldSlug", taskSlug).Str("newSlug", slug).Str("fileName", task.OriginalName).Msg("retry-task")
	return &InitResponse{
		UploadSlug:      slug,
		TotalChunks:     totalChunks,
		ChunkSize:       s.cfg.Upload.ChunkSize,
		CompletedChunks: []int{},
	}, nil
}

func (s *UploadService) ensureUploadParentUnlocked(ctx context.Context, userID int64, sessionID, parentSlug string) error {
	if parentSlug == "" {
		return nil
	}
	parent, err := s.queries.GetFileBySlugForUser(ctx, sqlc.GetFileBySlugForUserParams{
		Slug:   parentSlug,
		UserID: userID,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return model.ErrNotFound
		}
		return fmt.Errorf("get upload parent: %w", err)
	}
	if parent.IsTrashed || !parent.IsDir {
		return model.ErrNotFound
	}
	return ensureFileUnlocked(ctx, s.pg, userID, sessionID, parent.ID)
}
