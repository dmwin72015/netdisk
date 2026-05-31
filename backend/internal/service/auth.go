package service

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/crypto/bcrypt"

	"github.com/netdisk/server/internal/config"
	"github.com/netdisk/server/internal/db/sqlc"
	"github.com/netdisk/server/internal/model"
	"github.com/netdisk/server/pkg/jwtutil"
	gonanoid "github.com/matoous/go-nanoid/v2"
)

type AuthService struct {
	queries *sqlc.Queries
	pg      *pgxpool.Pool
	jwtMgr  *jwtutil.Manager
	cfg     *config.Config
}

func NewAuthService(queries *sqlc.Queries, pg *pgxpool.Pool, jwtMgr *jwtutil.Manager, cfg *config.Config) *AuthService {
	return &AuthService{queries: queries, pg: pg, jwtMgr: jwtMgr, cfg: cfg}
}

type RegisterInput struct {
	Username string `json:"username"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

type LoginInput struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type RefreshInput struct {
	RefreshToken string `json:"refreshToken"`
}

type LogoutInput struct {
	RefreshToken string `json:"refreshToken"`
}

type TokenPair struct {
	AccessToken  string `json:"accessToken"`
	RefreshToken string `json:"refreshToken"`
	ExpiresIn    int64  `json:"expiresIn"`
}

type UserResponse struct {
	Slug      string      `json:"slug"`
	Username  string      `json:"username"`
	Email     string      `json:"email"`
	Status    int16       `json:"status"`
	Profile   ProfileData `json:"profile"`
	Storage   StorageData `json:"storage"`
	Level     LevelData   `json:"level"`
	CreatedAt string      `json:"createdAt"`
}

func (s *AuthService) Register(ctx context.Context, input RegisterInput) (*UserResponse, error) {
	if input.Username == "" || input.Email == "" || input.Password == "" {
		return nil, model.ErrInvalidInput
	}
	if len(input.Password) < 6 {
		return nil, model.ErrInvalidInput
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(input.Password), s.cfg.Limits.BcryptCost)
	if err != nil {
		return nil, fmt.Errorf("hash password: %w", err)
	}

	slug, err := gonanoid.New(21)
	if err != nil {
		return nil, fmt.Errorf("generate slug: %w", err)
	}

	tx, err := s.pg.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback(ctx)

	qtx := s.queries.WithTx(tx)

	user, err := qtx.CreateUser(ctx, sqlc.CreateUserParams{
		Slug:         slug,
		Username:     input.Username,
		Email:        input.Email,
		PasswordHash: string(hash),
	})
	if err != nil {
		if isUniqueViolation(err) {
			return nil, model.ErrAlreadyExists
		}
		return nil, fmt.Errorf("create user: %w", err)
	}

	displayName := input.Username
	_, err = qtx.CreateProfile(ctx, sqlc.CreateProfileParams{
		UserID:      user.ID,
		DisplayName: pgtypeText(displayName),
	})
	if err != nil {
		return nil, fmt.Errorf("create profile: %w", err)
	}

	_, err = qtx.CreateStorageStats(ctx, sqlc.CreateStorageStatsParams{
		UserID:       user.ID,
		StorageQuota: s.cfg.Limits.DefaultStorageQuota,
	})
	if err != nil {
		return nil, fmt.Errorf("create storage stats: %w", err)
	}

	_, err = qtx.CreateLevel(ctx, sqlc.CreateLevelParams{
		UserID:    user.ID,
		LevelCode: "vip1",
		LevelName: "VIP1",
	})
	if err != nil {
		return nil, fmt.Errorf("create level: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("commit: %w", err)
	}

	return &UserResponse{
		Slug:      user.Slug,
		Username:  user.Username,
		Email:     user.Email,
		Status:    user.Status,
		Profile: ProfileData{
			DisplayName: user.Username,
		},
		Storage: StorageData{
			StorageQuota: s.cfg.Limits.DefaultStorageQuota,
		},
		Level: LevelData{
			LevelCode: "vip1",
			LevelName: "VIP1",
		},
		CreatedAt: user.CreatedAt.Time.Format(time.RFC3339),
	}, nil
}

func (s *AuthService) Login(ctx context.Context, input LoginInput) (*UserResponse, *TokenPair, error) {
	if input.Email == "" || input.Password == "" {
		return nil, nil, model.ErrInvalidInput
	}

	user, err := s.queries.GetUserByEmail(ctx, input.Email)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil, model.ErrUnauthorized
		}
		return nil, nil, fmt.Errorf("get user: %w", err)
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(input.Password)); err != nil {
		return nil, nil, model.ErrUnauthorized
	}

	tokens, err := s.issueTokens(ctx, user.ID)
	if err != nil {
		return nil, nil, err
	}

	resp := &UserResponse{
		Slug:      user.Slug,
		Username:  user.Username,
		Email:     user.Email,
		Status:    user.Status,
		CreatedAt: user.CreatedAt.Time.Format(time.RFC3339),
	}

	// Fetch profile
	profile, err := s.queries.GetProfileByUserID(ctx, user.ID)
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return nil, nil, fmt.Errorf("get profile: %w", err)
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

	// Fetch storage stats
	stats, err := s.queries.GetStorageStats(ctx, user.ID)
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return nil, nil, fmt.Errorf("get storage stats: %w", err)
	}
	if stats.ID != 0 {
		resp.Storage = StorageData{
			StorageUsed:  stats.StorageUsed,
			StorageQuota: stats.StorageQuota,
		}
	}

	// Fetch level
	level, err := s.queries.GetLevelByUserID(ctx, user.ID)
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return nil, nil, fmt.Errorf("get level: %w", err)
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

	return resp, tokens, nil
}

func (s *AuthService) Refresh(ctx context.Context, input RefreshInput) (*TokenPair, error) {
	if input.RefreshToken == "" {
		return nil, model.ErrInvalidInput
	}

	claims, err := s.jwtMgr.Parse(input.RefreshToken)
	if err != nil {
		return nil, model.ErrUnauthorized
	}
	if claims.Type != jwtutil.TokenTypeRefresh {
		return nil, model.ErrUnauthorized
	}

	tokenHash := hashToken(input.RefreshToken)
	rt, err := s.queries.GetRefreshTokenByHash(ctx, tokenHash)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, model.ErrUnauthorized
		}
		return nil, fmt.Errorf("get refresh token: %w", err)
	}

	// Revoke the old token.
	if err := s.queries.RevokeRefreshToken(ctx, rt.ID); err != nil {
		return nil, fmt.Errorf("revoke token: %w", err)
	}

	// Issue new tokens.
	tokens, err := s.issueTokens(ctx, claims.UserID)
	if err != nil {
		return nil, err
	}

	return tokens, nil
}

func (s *AuthService) Logout(ctx context.Context, refreshToken string) error {
	if refreshToken == "" {
		return model.ErrInvalidInput
	}

	tokenHash := hashToken(refreshToken)
	rt, err := s.queries.GetRefreshTokenByHash(ctx, tokenHash)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil // Already invalid, treat as success.
		}
		return fmt.Errorf("get refresh token: %w", err)
	}

	return s.queries.RevokeRefreshToken(ctx, rt.ID)
}

func (s *AuthService) issueTokens(ctx context.Context, userID int64) (*TokenPair, error) {
	accessTok, _, err := s.jwtMgr.GenerateAccessToken(userID)
	if err != nil {
		return nil, fmt.Errorf("generate access token: %w", err)
	}

	refreshTok, _, expiresAt, err := s.jwtMgr.GenerateRefreshToken(userID)
	if err != nil {
		return nil, fmt.Errorf("generate refresh token: %w", err)
	}

	tokenHash := hashToken(refreshTok)
	_, err = s.queries.CreateRefreshToken(ctx, sqlc.CreateRefreshTokenParams{
		UserID:    userID,
		TokenHash: tokenHash,
		ExpiresAt: pgtypeTimestamptz(expiresAt),
	})
	if err != nil {
		return nil, fmt.Errorf("store refresh token: %w", err)
	}

	return &TokenPair{
		AccessToken:  accessTok,
		RefreshToken: refreshTok,
		ExpiresIn:    int64(s.jwtMgr.AccessTTL().Seconds()),
	}, nil
}

func hashToken(token string) string {
	h := sha256.Sum256([]byte(token))
	return hex.EncodeToString(h[:])
}

func isUniqueViolation(err error) bool {
	// pgx wraps PG errors; check for unique_violation SQLSTATE 23505
	return err != nil && (contains(err.Error(), "duplicate key") || contains(err.Error(), "23505"))
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsSubstr(s, substr))
}

func containsSubstr(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

func pgtypeText(s string) pgtype.Text {
	return pgtype.Text{String: s, Valid: true}
}

func pgtypeTimestamptz(t time.Time) pgtype.Timestamptz {
	return pgtype.Timestamptz{Time: t, Valid: true}
}
