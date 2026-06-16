package service

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/netdisk/server/internal/config"
	"github.com/netdisk/server/internal/db/sqlc"
	"github.com/netdisk/server/internal/model"
	"github.com/netdisk/server/internal/storage"
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

type OAuthAccountInfo struct {
	Provider          string `json:"provider"`
	ProviderAccountID string `json:"providerAccountId"`
	OAuthEmail        string `json:"oauthEmail,omitempty"`
	CreatedAt         string `json:"createdAt"`
}

type UserMeResponse struct {
	Slug           string             `json:"slug"`
	Username       string             `json:"username"`
	Email          string             `json:"email"`
	Status         int16              `json:"status"`
	Role           string             `json:"role"`
	RegisterMethod string             `json:"registerMethod"`
	Profile        ProfileData        `json:"profile"`
	Storage        StorageData        `json:"storage"`
	Level          LevelData          `json:"level"`
	OAuthAccounts  []OAuthAccountInfo `json:"oauthAccounts"`
	CreatedAt      string             `json:"createdAt"`
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

	oauthInfos := s.getOAuthAccounts(ctx, userID)

	resp := &UserMeResponse{
		Slug:           user.Slug,
		Username:       user.Username,
		Email:          user.Email,
		Status:         user.Status,
		Role:           user.Role,
		RegisterMethod: user.RegisterMethod,
		OAuthAccounts:  oauthInfos,
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

func (s *UserService) getOAuthAccounts(ctx context.Context, userID int64) []OAuthAccountInfo {
	rows, err := s.pg.Query(ctx, `
		SELECT provider, provider_account_id, COALESCE(oauth_email, ''), created_at
		FROM user_oauth_accounts
		WHERE user_id = $1
		ORDER BY created_at ASC
	`, userID)
	if err != nil {
		return nil
	}
	defer rows.Close()

	var infos []OAuthAccountInfo
	for rows.Next() {
		var provider, providerAccountID, oauthEmail string
		var createdAt time.Time
		if err := rows.Scan(&provider, &providerAccountID, &oauthEmail, &createdAt); err != nil {
			continue
		}
		infos = append(infos, OAuthAccountInfo{
			Provider:          provider,
			ProviderAccountID: providerAccountID,
			OAuthEmail:        oauthEmail,
			CreatedAt:         createdAt.Format(time.RFC3339),
		})
	}
	return infos
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

	if _, err := s.queries.GetUserByID(ctx, userID); err != nil {
		return "", fmt.Errorf("get user: %w", err)
	}

	tmpDir := filepath.Join(s.cfg.Storage.Root, s.cfg.Storage.TmpDir)
	if err := os.MkdirAll(tmpDir, 0o755); err != nil {
		return "", fmt.Errorf("create avatar temp dir: %w", err)
	}

	tmpFile, err := os.CreateTemp(tmpDir, "avatar-*")
	if err != nil {
		return "", fmt.Errorf("create avatar temp file: %w", err)
	}
	tmpPath := tmpFile.Name()
	keepTemp := false
	defer func() {
		if !keepTemp {
			_ = os.Remove(tmpPath)
		}
	}()

	ext := ".jpg"
	switch mimeType {
	case "image/png":
		ext = ".png"
	case "image/webp":
		ext = ".webp"
	}

	hash := sha256.New()
	written, err := io.Copy(io.MultiWriter(tmpFile, hash), limitedReader)
	if err != nil {
		_ = tmpFile.Close()
		return "", fmt.Errorf("write avatar: %w", err)
	}
	if err := tmpFile.Close(); err != nil {
		return "", fmt.Errorf("close avatar temp file: %w", err)
	}
	if written > maxAvatarSize {
		return "", model.ErrFileTooLarge
	}
	if err := os.Chmod(tmpPath, 0o644); err != nil {
		return "", fmt.Errorf("chmod avatar temp file: %w", err)
	}

	avatarHash := hex.EncodeToString(hash.Sum(nil))
	avatarRelPath := storage.StoragePath(avatarHash, s.cfg.Storage.AvatarsDir) + ext
	avatarPath := filepath.Join(s.cfg.Storage.Root, avatarRelPath)
	if err := os.MkdirAll(filepath.Dir(avatarPath), 0o755); err != nil {
		return "", fmt.Errorf("create avatar dir: %w", err)
	}

	if _, err := os.Stat(avatarPath); err == nil {
		return path.Join("/api/v1", filepath.ToSlash(avatarRelPath)), nil
	} else if !os.IsNotExist(err) {
		return "", fmt.Errorf("stat avatar: %w", err)
	}

	if err := os.Rename(tmpPath, avatarPath); err != nil {
		if _, statErr := os.Stat(avatarPath); statErr == nil {
			return path.Join("/api/v1", filepath.ToSlash(avatarRelPath)), nil
		}
		return "", fmt.Errorf("move avatar: %w", err)
	}
	keepTemp = true

	return path.Join("/api/v1", filepath.ToSlash(avatarRelPath)), nil
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
