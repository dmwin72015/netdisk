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

	"github.com/netdisk/server/internal/cache"
	"github.com/netdisk/server/internal/config"
	"github.com/netdisk/server/internal/db/sqlc"
	"github.com/netdisk/server/internal/model"
	"github.com/netdisk/server/internal/storage"
)

const (
	chunkSize = 4 * 1024 * 1024 // 4 MB
)

type UploadService struct {
	queries *sqlc.Queries
	pg      *pgxpool.Pool
	cfg     *config.Config
	store   *storage.Local
	cache   *cache.Cache
}

func NewUploadService(queries *sqlc.Queries, pg *pgxpool.Pool, cfg *config.Config, store *storage.Local, c *cache.Cache) *UploadService {
	return &UploadService{queries: queries, pg: pg, cfg: cfg, store: store, cache: c}
}

type PreCheckRequest struct {
	PreHash  string `json:"pre_hash"`
	FileSize int64  `json:"file_size"`
}

type PreCheckResponse struct {
	Status string `json:"status"`
}

type RequestChallengeRequest struct {
	FileHash string `json:"file_hash"`
}

type RequestChallengeResponse struct {
	Status          string `json:"status"`
	ChallengeOffset int64  `json:"challenge_offset,omitempty"`
	ChallengeToken  string `json:"challenge_token,omitempty"`
}

type VerifyRequest struct {
	FileHash  string `json:"file_hash"`
	ProofCode string `json:"proof_code"`
}

type VerifyResponse struct {
	Status           string `json:"status"`
	PhysicalFileSlug string `json:"physical_file_slug,omitempty"`
}

type InitRequest struct {
	FileHash string `json:"file_hash"`
	PreHash  string `json:"pre_hash"`
	FileSize int64  `json:"file_size"`
	MimeType string `json:"mime_type"`
}

type InitResponse struct {
	UploadSlug      string `json:"upload_slug"`
	TotalChunks     int32  `json:"total_chunks"`
	ChunkSize       int32  `json:"chunk_size"`
	CompletedChunks []int  `json:"completed_chunks"`
}

type CompleteResponse struct {
	Status           string `json:"status"`
	PhysicalFileSlug string `json:"physical_file_slug,omitempty"`
}

type StatusResponse struct {
	Status           string  `json:"status"`
	PhysicalFileSlug string  `json:"physical_file_slug,omitempty"`
	Error            *string `json:"error,omitempty"`
}

func (s *UploadService) PreCheck(ctx context.Context, userID int64, req PreCheckRequest) (*PreCheckResponse, error) {
	if req.FileSize <= 0 {
		return nil, model.ErrInvalidInput
	}
	if req.FileSize > s.cfg.Storage.MaxUploadSize {
		return nil, model.ErrQuotaExceeded
	}

	// Check Redis pre-cache
	_, err := s.cache.PreCache.Get(ctx, req.FileSize, req.PreHash)
	if err == nil {
		return &PreCheckResponse{Status: "SUSPECT_HIT"}, nil
	}

	// Check DB
	_, err = s.queries.GetPhysicalFileByPreHash(ctx, sqlc.GetPhysicalFileByPreHashParams{
		PreHash:  req.PreHash,
		FileSize: req.FileSize,
	})
	if err == nil {
		_ = s.cache.PreCache.Set(ctx, req.FileSize, req.PreHash, "")
		return &PreCheckResponse{Status: "SUSPECT_HIT"}, nil
	}

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
			return &RequestChallengeResponse{Status: "NOT_FOUND"}, nil
		}
		return nil, fmt.Errorf("get physical file: %w", err)
	}

	offset := int(rand.Int63n(max(1, pf.FileSize-1024)))
	token, err := gonanoid.New(32)
	if err != nil {
		return nil, fmt.Errorf("generate token: %w", err)
	}

	if err := s.cache.Challenge.SetChallenge(ctx, userID, req.FileHash, offset, token); err != nil {
		return nil, fmt.Errorf("set challenge: %w", err)
	}

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
		return nil, model.ErrChallengeExpired
	}

	diskBytes, err := s.store.ReadAt(req.FileHash, int64(offset), 1024)
	if err != nil {
		return nil, fmt.Errorf("read file: %w", err)
	}

	h := sha256.New()
	h.Write(diskBytes)
	h.Write([]byte(token))
	expected := hex.EncodeToString(h.Sum(nil))

	if expected != req.ProofCode {
		return &VerifyResponse{Status: "MISS"}, nil
	}

	pf, err := s.queries.GetPhysicalFileByHash(ctx, sqlc.GetPhysicalFileByHashParams{
		HashAlgo: "sha256",
		FileHash: req.FileHash,
	})
	if err != nil {
		return nil, fmt.Errorf("get physical file: %w", err)
	}

	return &VerifyResponse{
		Status:           "HIT",
		PhysicalFileSlug: pf.Slug,
	}, nil
}

func (s *UploadService) Init(ctx context.Context, userID int64, req InitRequest) (*InitResponse, error) {
	if req.FileHash == "" || req.PreHash == "" || req.FileSize <= 0 || req.MimeType == "" {
		return nil, model.ErrInvalidInput
	}
	if req.FileSize > s.cfg.Storage.MaxUploadSize {
		return nil, model.ErrQuotaExceeded
	}

	existing, err := s.queries.GetUploadTaskByHashForUser(ctx, sqlc.GetUploadTaskByHashForUserParams{
		OwnerUserID: userID,
		FileHash:    req.FileHash,
	})
	if err == nil && existing.ID != 0 && existing.Status != "done" && existing.Status != "failed" {
		chunks, _ := s.cache.Chunks.ListChunks(ctx, existing.Slug)
		return &InitResponse{
			UploadSlug:      existing.Slug,
			TotalChunks:     existing.TotalChunks,
			ChunkSize:       existing.ChunkSize,
			CompletedChunks: chunks,
		}, nil
	}

	totalChunks := int32(math.Ceil(float64(req.FileSize) / float64(chunkSize)))

	slug, err := gonanoid.New(21)
	if err != nil {
		return nil, fmt.Errorf("generate slug: %w", err)
	}

	_, err = s.queries.CreateUploadTask(ctx, sqlc.CreateUploadTaskParams{
		Slug:        slug,
		OwnerUserID: userID,
		HashAlgo:    "sha256",
		FileHash:    req.FileHash,
		PreHash:     req.PreHash,
		FileSize:    req.FileSize,
		MimeType:    req.MimeType,
		TotalChunks: totalChunks,
		ChunkSize:   chunkSize,
		Status:      "uploading",
		ExpiresAt:   pgtype.Timestamptz{Time: time.Now().Add(24 * 30 * time.Hour), Valid: true},
	})
	if err != nil {
		return nil, fmt.Errorf("create upload task: %w", err)
	}

	return &InitResponse{
		UploadSlug:      slug,
		TotalChunks:     totalChunks,
		ChunkSize:       chunkSize,
		CompletedChunks: []int{},
	}, nil
}

func (s *UploadService) AppendChunk(ctx context.Context, userID int64, uploadSlug string, chunkIndex int32, data []byte) error {
	task, err := s.queries.GetUploadTaskBySlug(ctx, uploadSlug)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return model.ErrNotFound
		}
		return fmt.Errorf("get task: %w", err)
	}

	if task.OwnerUserID != userID {
		return model.ErrUnauthorized
	}
	if task.Status != "uploading" {
		return model.ErrInvalidInput
	}
	if chunkIndex < 0 || chunkIndex >= task.TotalChunks {
		return model.ErrInvalidInput
	}

	isLast := chunkIndex == task.TotalChunks-1
	expectedSize := int(task.ChunkSize)
	if isLast {
		remaining := int(task.FileSize) - int(chunkIndex)*int(task.ChunkSize)
		expectedSize = remaining
	}
	if len(data) != expectedSize {
		return model.ErrInvalidInput
	}

	if err := s.store.WriteChunk(uploadSlug, int(chunkIndex), bytes.NewReader(data)); err != nil {
		return fmt.Errorf("write chunk: %w", err)
	}

	if err := s.cache.Chunks.AddChunk(ctx, uploadSlug, int(chunkIndex)); err != nil {
		return fmt.Errorf("track chunk: %w", err)
	}

	return nil
}

func (s *UploadService) Complete(ctx context.Context, userID int64, uploadSlug string) (*CompleteResponse, error) {
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
	if task.Status != "uploading" {
		return nil, model.ErrInvalidInput
	}

	if err := s.queries.UpdateUploadTaskStatus(ctx, sqlc.UpdateUploadTaskStatusParams{
		ID:     task.ID,
		Status: "merging",
	}); err != nil {
		return nil, fmt.Errorf("update status: %w", err)
	}

	acquired, err := s.cache.Lock.AcquireMergeLock(ctx, task.FileHash)
	if err != nil {
		return nil, fmt.Errorf("acquire lock: %w", err)
	}
	if !acquired {
		return &CompleteResponse{Status: "MERGING"}, nil
	}
	defer s.cache.Lock.ReleaseMergeLock(ctx, task.FileHash)

	_, err = s.store.MergeChunks(uploadSlug, task.FileHash, int(task.TotalChunks))
	if err != nil {
		_ = s.queries.UpdateUploadTaskStatus(ctx, sqlc.UpdateUploadTaskStatusParams{
			ID:     task.ID,
			Status: "failed",
		})
		return nil, fmt.Errorf("merge chunks: %w", err)
	}

	storagePath := storage.StoragePath(task.FileHash)
	pf, err := s.queries.CreatePhysicalFile(ctx, sqlc.CreatePhysicalFileParams{
		Slug:        task.Slug,
		HashAlgo:    "sha256",
		FileHash:    task.FileHash,
		PreHash:     task.PreHash,
		FileSize:    task.FileSize,
		MimeType:    task.MimeType,
		StoragePath: storagePath,
		Status:      "completed",
	})
	if err != nil {
		pf, err = s.queries.GetPhysicalFileByHash(ctx, sqlc.GetPhysicalFileByHashParams{
			HashAlgo: "sha256",
			FileHash: task.FileHash,
		})
		if err != nil {
			return nil, fmt.Errorf("get physical file: %w", err)
		}
	}

	_ = s.queries.UpdateUploadTaskPhysicalFile(ctx, sqlc.UpdateUploadTaskPhysicalFileParams{
		ID:             task.ID,
		PhysicalFileID: pgtype.Int8{Int64: pf.ID, Valid: true},
	})
	_ = s.queries.UpdateUploadTaskStatus(ctx, sqlc.UpdateUploadTaskStatusParams{
		ID:     task.ID,
		Status: "done",
	})

	_ = s.cache.PreCache.Set(ctx, task.FileSize, task.PreHash, pf.Slug)
	_ = s.store.CleanupUpload(uploadSlug)
	_ = s.cache.Chunks.DeleteChunks(ctx, uploadSlug)

	return &CompleteResponse{
		Status:           "DONE",
		PhysicalFileSlug: pf.Slug,
	}, nil
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

	return resp, nil
}
