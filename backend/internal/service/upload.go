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
	FileHash string `json:"fileHash"`
	PreHash  string `json:"preHash"`
	FileSize int64  `json:"fileSize"`
	MimeType string `json:"mimeType"`
	FileName string `json:"fileName"`
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

	return &VerifyResponse{
		Status:           "HIT",
		PhysicalFileSlug: pf.Slug,
		ExistingFiles:    existingFiles,
	}, nil
}

func (s *UploadService) Init(ctx context.Context, userID int64, req InitRequest) (*InitResponse, error) {
	if req.PreHash == "" || req.FileSize <= 0 || req.MimeType == "" {
		return nil, model.ErrInvalidInput
	}
	if req.FileSize > s.cfg.Storage.MaxUploadSize {
		return nil, model.ErrFileTooLarge
	}

	// Resume check only when hash is known
	if req.FileHash != "" {
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
	})
	if err != nil {
		return nil, fmt.Errorf("create upload task: %w", err)
	}

	return &InitResponse{
		UploadSlug:      slug,
		TotalChunks:     totalChunks,
		ChunkSize:       s.cfg.Upload.ChunkSize,
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
		return nil // already set
	}
	if err := s.queries.UpdateUploadTaskFileHash(ctx, sqlc.UpdateUploadTaskFileHashParams{
		ID:       task.ID,
		FileHash: req.FileHash,
	}); err != nil {
		return fmt.Errorf("update hash: %w", err)
	}
	if req.PreHash != "" {
		_ = s.queries.UpdateUploadTaskPreHash(ctx, sqlc.UpdateUploadTaskPreHashParams{
			ID:      task.ID,
			PreHash: req.PreHash,
		})
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

	fileHash := task.FileHash
	var storagePath string

	if fileHash == "" {
		// Hash not known yet — merge chunks and compute hash.
		computedHash, sp, err := s.store.MergeChunksAndHash(uploadSlug, int(task.TotalChunks))
		if err != nil {
			_ = s.queries.UpdateUploadTaskStatus(ctx, sqlc.UpdateUploadTaskStatusParams{ID: task.ID, Status: "failed"})
			return nil, fmt.Errorf("merge and hash: %w", err)
		}
		fileHash = computedHash
		storagePath = sp
		_ = s.queries.UpdateUploadTaskFileHash(ctx, sqlc.UpdateUploadTaskFileHashParams{ID: task.ID, FileHash: fileHash})
	} else {
		acquired, err := s.cache.Lock.AcquireMergeLock(ctx, fileHash)
		if err != nil {
			return nil, fmt.Errorf("acquire lock: %w", err)
		}
		if !acquired {
			return &CompleteResponse{Status: "MERGING"}, nil
		}
		defer s.cache.Lock.ReleaseMergeLock(ctx, fileHash)

		sp, err := s.store.MergeChunks(uploadSlug, fileHash, int(task.TotalChunks))
		if err != nil {
			_ = s.queries.UpdateUploadTaskStatus(ctx, sqlc.UpdateUploadTaskStatusParams{ID: task.ID, Status: "failed"})
			return nil, fmt.Errorf("merge chunks: %w", err)
		}
		storagePath = sp
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
		pf, err = s.queries.GetPhysicalFileByHash(ctx, sqlc.GetPhysicalFileByHashParams{
			HashAlgo: "sha256",
			FileHash: fileHash,
		})
		if err != nil {
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

type TaskItem struct {
	Slug          string `json:"slug"`
	FileName      string `json:"fileName"`
	FileSize      int64  `json:"fileSize"`
	MimeType      string `json:"mimeType"`
	Status        string `json:"status"`
	ErrorMsg      string `json:"errorMsg,omitempty"`
	TotalChunks   int32  `json:"totalChunks"`
	ReceivedBytes int64  `json:"receivedBytes"`
	CreatedAt     string `json:"createdAt"`
	UpdatedAt     string `json:"updatedAt"`
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

	// Clean up Redis chunk tracking
	_ = s.cache.Chunks.DeleteChunks(ctx, taskSlug)

	// Delete the task row
	if err := s.queries.DeleteUploadTaskBySlug(ctx, sqlc.DeleteUploadTaskBySlugParams{
		Slug:        taskSlug,
		OwnerUserID: userID,
	}); err != nil {
		return fmt.Errorf("delete task: %w", err)
	}

	return nil
}

func (s *UploadService) DeleteTasks(ctx context.Context, userID int64, slugs []string) error {
	if len(slugs) == 0 {
		return nil
	}

	// Clean up Redis chunk tracking for each task
	for _, slug := range slugs {
		_ = s.cache.Chunks.DeleteChunks(ctx, slug)
	}

	// Batch delete
	if err := s.queries.DeleteUploadTasksBySlugs(ctx, sqlc.DeleteUploadTasksBySlugsParams{
		Column1:     slugs,
		OwnerUserID: userID,
	}); err != nil {
		return fmt.Errorf("delete tasks: %w", err)
	}

	return nil
}

func (s *UploadService) RetryTask(ctx context.Context, userID int64, taskSlug string) (*InitResponse, error) {
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
		return nil, model.ErrInvalidInput
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
	})
	if err != nil {
		return nil, fmt.Errorf("create upload task: %w", err)
	}

	return &InitResponse{
		UploadSlug:      slug,
		TotalChunks:     totalChunks,
		ChunkSize:       s.cfg.Upload.ChunkSize,
		CompletedChunks: []int{},
	}, nil
}
