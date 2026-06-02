package service

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/netdisk/server/internal/config"
	"github.com/netdisk/server/internal/db/sqlc"
	"github.com/netdisk/server/internal/model"
	"golang.org/x/crypto/bcrypt"
)

type UserService struct {
	queries *sqlc.Queries
	pg      *pgxpool.Pool
	cfg     *config.Config
}

func NewUserService(queries *sqlc.Queries, pg *pgxpool.Pool, cfg *config.Config) *UserService {
	return &UserService{queries: queries, pg: pg, cfg: cfg}
}

type UserMeResponse struct {
	Slug           string      `json:"slug"`
	Username       string      `json:"username"`
	Email          string      `json:"email"`
	Status         int16       `json:"status"`
	Role           string      `json:"role"`
	RegisterMethod string      `json:"registerMethod"`
	Profile        ProfileData `json:"profile"`
	Storage        StorageData `json:"storage"`
	Level          LevelData   `json:"level"`
	CreatedAt      string      `json:"createdAt"`
}

type ProfileData struct {
	DisplayName string `json:"displayName"`
	AvatarURL   string `json:"avatarUrl"`
	Bio         string `json:"bio"`
}

type StorageData struct {
	StorageUsed  int64 `json:"storageUsed"`
	StorageQuota int64 `json:"storageQuota"`
}

type LevelData struct {
	LevelCode string  `json:"levelCode"`
	LevelName string  `json:"levelName"`
	ExpiresAt *string `json:"expiresAt"`
}

func (s *UserService) GetMe(ctx context.Context, userID int64) (*UserMeResponse, error) {
	user, err := s.queries.GetUserByID(ctx, userID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, model.ErrNotFound
		}
		return nil, fmt.Errorf("get user: %w", err)
	}

	profile, err := s.queries.GetProfileByUserID(ctx, userID)
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return nil, fmt.Errorf("get profile: %w", err)
	}

	stats, err := s.queries.GetStorageStats(ctx, userID)
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return nil, fmt.Errorf("get storage stats: %w", err)
	}

	level, err := s.queries.GetLevelByUserID(ctx, userID)
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return nil, fmt.Errorf("get level: %w", err)
	}

	resp := &UserMeResponse{
		Slug:           user.Slug,
		Username:       user.Username,
		Email:          user.Email,
		Status:         user.Status,
		Role:           user.Role,
		RegisterMethod: user.RegisterMethod,
		CreatedAt:      user.CreatedAt.Time.Format("2006-01-02T15:04:05Z"),
	}

	if profile.ID != 0 {
		resp.Profile = ProfileData{
			DisplayName: profile.DisplayName.String,
			AvatarURL:   profile.AvatarPath.String,
			Bio:         profile.Bio.String,
		}
	} else {
		resp.Profile = ProfileData{
			DisplayName: user.Username,
		}
	}

	if stats.ID != 0 {
		resp.Storage = StorageData{
			StorageUsed:  stats.StorageUsed,
			StorageQuota: stats.StorageQuota,
		}
	}

	if level.ID != 0 {
		ld := LevelData{
			LevelCode: level.LevelCode,
			LevelName: level.LevelName,
		}
		if level.ExpiresAt.Valid {
			s := level.ExpiresAt.Time.Format("2006-01-02T15:04:05Z")
			ld.ExpiresAt = &s
		}
		resp.Level = ld
	}

	return resp, nil
}

type CategoryStat struct {
	Category string `json:"category"`
	Bytes    int64  `json:"bytes"`
	Count    int64  `json:"count"`
}

func (s *UserService) GetStorageBreakdown(ctx context.Context, userID int64) ([]CategoryStat, error) {
	rows, err := s.pg.Query(ctx,
		`SELECT COALESCE(NULLIF(file_category, ''), 'other'), COALESCE(SUM(file_size), 0), COUNT(*)
		 FROM user_files
		 WHERE user_id = $1 AND is_dir = FALSE AND is_trashed = FALSE
		 GROUP BY COALESCE(NULLIF(file_category, ''), 'other')
		 ORDER BY SUM(file_size) DESC`,
		userID,
	)
	if err != nil {
		return nil, fmt.Errorf("query storage breakdown: %w", err)
	}
	defer rows.Close()

	statMap := make(map[string]CategoryStat, 8)
	for rows.Next() {
		var cs CategoryStat
		if err := rows.Scan(&cs.Category, &cs.Bytes, &cs.Count); err != nil {
			return nil, fmt.Errorf("scan category stat: %w", err)
		}
		statMap[cs.Category] = cs
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	// Query trashed files separately
	var trashBytes, trashCount int64
	_ = s.pg.QueryRow(ctx,
		`SELECT COALESCE(SUM(file_size), 0), COUNT(*)
		 FROM user_files
		 WHERE user_id = $1 AND is_dir = FALSE AND is_trashed = TRUE`,
		userID,
	).Scan(&trashBytes, &trashCount)

	allCategories := []string{"video", "audio", "image", "document", "archive", "other", "trash"}
	stats := make([]CategoryStat, 0, len(allCategories))
	for _, cat := range allCategories {
		var cs CategoryStat
		if cat == "trash" {
			cs = CategoryStat{Category: "trash", Bytes: trashBytes, Count: trashCount}
		} else if c, ok := statMap[cat]; ok {
			cs = c
		} else {
			cs = CategoryStat{Category: cat}
		}
		stats = append(stats, cs)
	}

	return stats, nil
}

func (s *UserService) UpdateProfile(ctx context.Context, userID int64, displayName, bio, avatarURL string) error {
	if err := s.queries.UpdateProfile(ctx, sqlc.UpdateProfileParams{
		UserID:      userID,
		DisplayName: pgtype.Text{String: displayName, Valid: true},
		Bio:         pgtype.Text{String: bio, Valid: true},
	}); err != nil {
		return err
	}
	if avatarURL != "" {
		if err := s.queries.UpdateAvatarPath(ctx, sqlc.UpdateAvatarPathParams{
			UserID:     userID,
			AvatarPath: pgtype.Text{String: avatarURL, Valid: true},
		}); err != nil {
			return err
		}
	}
	return nil
}

func (s *UserService) ChangePassword(ctx context.Context, userID int64, oldPassword, newPassword string) error {
	if newPassword == "" || len(newPassword) < 6 {
		return model.ErrInvalidInput
	}

	user, err := s.queries.GetUserByID(ctx, userID)
	if err != nil {
		return fmt.Errorf("get user: %w", err)
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(oldPassword)); err != nil {
		return model.ErrUnauthorized
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(newPassword), s.cfg.Limits.BcryptCost)
	if err != nil {
		return fmt.Errorf("hash password: %w", err)
	}

	// Direct update since there's no UpdatePassword query - use raw SQL via pool
	_, err = s.pg.Exec(ctx,
		"UPDATE users SET password_hash = $1, updated_at = NOW() WHERE id = $2",
		string(hash), userID)
	if err != nil {
		return fmt.Errorf("update password: %w", err)
	}

	return nil
}

func (s *UserService) UploadAvatar(ctx context.Context, userID int64, reader io.Reader, mimeType string) (string, error) {
	if mimeType != "image/jpeg" && mimeType != "image/png" && mimeType != "image/webp" {
		return "", model.ErrUnsupportedType
	}

	// Limit avatar size from config
	maxAvatarSize := s.cfg.Limits.AvatarMaxSize
	limitedReader := io.LimitReader(reader, maxAvatarSize+1)

	user, err := s.queries.GetUserByID(ctx, userID)
	if err != nil {
		return "", fmt.Errorf("get user: %w", err)
	}

	ext := ".jpg"
	switch mimeType {
	case "image/png":
		ext = ".png"
	case "image/webp":
		ext = ".webp"
	}

	// Determine next version from current avatar path (/api/v1/avatars/{slug}/2.jpg)
	version := 1
	profile, err := s.queries.GetProfileByUserID(ctx, userID)
	if err == nil && profile.AvatarPath.Valid && profile.AvatarPath.String != "" {
		raw := strings.TrimSuffix(filepath.Base(profile.AvatarPath.String), filepath.Ext(profile.AvatarPath.String))
		if v, err := strconv.Atoi(raw); err == nil {
			version = v + 1
		}
	}

	userDir := filepath.Join(s.cfg.Storage.Root, s.cfg.Storage.AvatarsDir, user.Slug)
	if err := os.MkdirAll(userDir, 0o755); err != nil {
		return "", fmt.Errorf("create avatar dir: %w", err)
	}

	filename := fmt.Sprintf("%d%s", version, ext)
	path := filepath.Join(userDir, filename)

	f, err := os.Create(path)
	if err != nil {
		return "", fmt.Errorf("create avatar: %w", err)
	}
	defer f.Close()

	written, err := io.Copy(f, limitedReader)
	if err != nil {
		return "", fmt.Errorf("write avatar: %w", err)
	}
	if written > maxAvatarSize {
		os.Remove(path)
		return "", model.ErrFileTooLarge
	}

	avatarURL := fmt.Sprintf("/api/v1/avatars/%s/%s", user.Slug, filename)
	return avatarURL, nil
}

func (s *UserService) ListTransactions(ctx context.Context, userID int64, page, pageSize int) ([]sqlc.UserTransaction, int, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}
	offset := (page - 1) * pageSize

	total, err := s.queries.CountTransactionsByUser(ctx, userID)
	if err != nil {
		return nil, 0, fmt.Errorf("count transactions: %w", err)
	}

	txs, err := s.queries.ListTransactionsByUser(ctx, sqlc.ListTransactionsByUserParams{
		UserID: userID,
		Limit:  int32(pageSize),
		Offset: int32(offset),
	})
	if err != nil {
		return nil, 0, fmt.Errorf("list transactions: %w", err)
	}

	return txs, int(total), nil
}
