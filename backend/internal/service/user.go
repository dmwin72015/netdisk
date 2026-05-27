package service

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
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
	Slug      string         `json:"slug"`
	Username  string         `json:"username"`
	Email     string         `json:"email"`
	Status    int16          `json:"status"`
	Profile   ProfileData    `json:"profile"`
	Storage   StorageData    `json:"storage"`
	Level     LevelData      `json:"level"`
	CreatedAt string         `json:"created_at"`
}

type ProfileData struct {
	DisplayName string `json:"display_name"`
	AvatarURL   string `json:"avatar_url"`
	Bio         string `json:"bio"`
}

type StorageData struct {
	StorageUsed  int64 `json:"storage_used"`
	StorageQuota int64 `json:"storage_quota"`
}

type LevelData struct {
	LevelCode string  `json:"level_code"`
	LevelName string  `json:"level_name"`
	ExpiresAt *string `json:"expires_at"`
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
		Slug:      user.Slug,
		Username:  user.Username,
		Email:     user.Email,
		Status:    user.Status,
		CreatedAt: user.CreatedAt.Time.Format("2006-01-02T15:04:05Z"),
	}

	if profile.ID != 0 {
		resp.Profile = ProfileData{
			DisplayName: profile.DisplayName.String,
			AvatarURL:   fmt.Sprintf("/api/v1/user/avatar/%s", user.Slug),
			Bio:         profile.Bio.String,
		}
	} else {
		resp.Profile = ProfileData{
			DisplayName: user.Username,
			AvatarURL:   fmt.Sprintf("/api/v1/user/avatar/%s", user.Slug),
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

func (s *UserService) UpdateProfile(ctx context.Context, userID int64, displayName, bio string) error {
	return s.queries.UpdateProfile(ctx, sqlc.UpdateProfileParams{
		UserID:      userID,
		DisplayName: pgtype.Text{String: displayName, Valid: true},
		Bio:         pgtype.Text{String: bio, Valid: true},
	})
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

	hash, err := bcrypt.GenerateFromPassword([]byte(newPassword), 12)
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
	if mimeType != "image/jpeg" && mimeType != "image/png" {
		return "", model.ErrUnsupportedType
	}

	user, err := s.queries.GetUserByID(ctx, userID)
	if err != nil {
		return "", fmt.Errorf("get user: %w", err)
	}

	ext := ".jpg"
	if mimeType == "image/png" {
		ext = ".png"
	}

	avatarDir := filepath.Join(s.cfg.Storage.Root, "avatars")
	filename := user.Slug + ext
	path := filepath.Join(avatarDir, filename)

	f, err := os.Create(path)
	if err != nil {
		return "", fmt.Errorf("create avatar: %w", err)
	}
	defer f.Close()

	if _, err := io.Copy(f, reader); err != nil {
		return "", fmt.Errorf("write avatar: %w", err)
	}

	avatarPath := "/avatars/" + filename
	if err := s.queries.UpdateAvatarPath(ctx, sqlc.UpdateAvatarPathParams{
		UserID:     userID,
		AvatarPath: pgtype.Text{String: avatarPath, Valid: true},
	}); err != nil {
		return "", fmt.Errorf("update avatar path: %w", err)
	}

	return fmt.Sprintf("/api/v1/user/avatar/%s", user.Slug), nil
}

func (s *UserService) GetAvatarPath(ctx context.Context, userSlug string) (string, error) {
	user, err := s.queries.GetUserBySlug(ctx, userSlug)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return "", model.ErrNotFound
		}
		return "", fmt.Errorf("get user: %w", err)
	}

	profile, err := s.queries.GetProfileByUserID(ctx, user.ID)
	if err != nil || !profile.AvatarPath.Valid || profile.AvatarPath.String == "" {
		return "", model.ErrNotFound
	}

	// Resolve to absolute path
	avatarPath := strings.TrimPrefix(profile.AvatarPath.String, "/avatars/")
	return filepath.Join(s.cfg.Storage.Root, "avatars", avatarPath), nil
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
