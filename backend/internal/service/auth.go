package service

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
	"golang.org/x/crypto/bcrypt"

	gonanoid "github.com/matoous/go-nanoid/v2"
	"github.com/netdisk/server/internal/config"
	"github.com/netdisk/server/internal/db/sqlc"
	"github.com/netdisk/server/internal/model"
	"github.com/netdisk/server/pkg/jwtutil"
)

type AuthService struct {
	queries *sqlc.Queries
	pg      *pgxpool.Pool
	jwtMgr  *jwtutil.Manager
	cfg     *config.Config
	rdb     *redis.Client
}

func NewAuthService(queries *sqlc.Queries, pg *pgxpool.Pool, jwtMgr *jwtutil.Manager, cfg *config.Config, rdb *redis.Client) *AuthService {
	return &AuthService{queries: queries, pg: pg, jwtMgr: jwtMgr, cfg: cfg, rdb: rdb}
}

func (s *AuthService) Cfg() *config.Config {
	return s.cfg
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
		Slug:           user.Slug,
		Username:       user.Username,
		Email:          user.Email,
		Status:         user.Status,
		Role:           user.Role,
		RegisterMethod: user.RegisterMethod,
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
		Slug:           user.Slug,
		Username:       user.Username,
		Email:          user.Email,
		Status:         user.Status,
		Role:           user.Role,
		RegisterMethod: user.RegisterMethod,
		CreatedAt:      user.CreatedAt.Time.Format(time.RFC3339),
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
	tokens, err := s.issueTokens(ctx, claims.UserID, claims.SID)
	if err != nil {
		return nil, err
	}

	return tokens, nil
}

func (s *AuthService) Logout(ctx context.Context, refreshToken string, userID int64, sessionID string) error {
	if refreshToken == "" {
		return model.ErrInvalidInput
	}

	tokenHash := hashToken(refreshToken)
	rt, err := s.queries.GetRefreshTokenByHash(ctx, tokenHash)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			s.cleanDirectoryUnlocks(ctx, userID, sessionID)
			return nil
		}
		return fmt.Errorf("get refresh token: %w", err)
	}

	err = s.queries.RevokeRefreshToken(ctx, rt.ID)

	// Always clean up directory unlocks for this session, even if revocation fails.
	s.cleanDirectoryUnlocks(ctx, userID, sessionID)

	return err
}

// cleanDirectoryUnlocks removes all directory unlocks for a user+session,
// and also cleans up any expired unlocks for the user.
func (s *AuthService) cleanDirectoryUnlocks(ctx context.Context, userID int64, sessionID string) {
	_, _ = s.pg.Exec(ctx,
		`DELETE FROM user_directory_unlocks
		 WHERE user_id = $1 AND session_id = $2`,
		userID, sessionID,
	)
	// Also clean up expired unlocks for the user.
	_, _ = s.pg.Exec(ctx,
		`DELETE FROM user_directory_unlocks
		 WHERE user_id = $1
		   AND expires_at IS NOT NULL
		   AND expires_at <= NOW()`,
		userID,
	)
}

func (s *AuthService) issueTokens(ctx context.Context, userID int64, sessionID ...string) (*TokenPair, error) {
	sid := ""
	if len(sessionID) > 0 {
		sid = sessionID[0]
	}
	if sid == "" {
		sid = makeState()
	}

	accessTok, _, err := s.jwtMgr.GenerateAccessToken(userID, sid)
	if err != nil {
		return nil, fmt.Errorf("generate access token: %w", err)
	}

	refreshTok, _, expiresAt, err := s.jwtMgr.GenerateRefreshToken(userID, sid)
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

// OAuth2 types.
type OAuthTokenResponse struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	ExpiresIn   int    `json:"expires_in"`
	Scope       string `json:"scope"`
}

type OAuthUserInfo struct {
	ID        string `json:"id"`
	Username  string `json:"username"`
	AvatarURL string `json:"avatar_url"`
}

// OAuthAuthorize returns the authorization URL for the given provider.
func (s *AuthService) OAuthAuthorize(ctx context.Context, provider string) (string, error) {
	prov, ok := s.cfg.OAuth2.Providers[provider]
	if !ok {
		return "", fmt.Errorf("unknown OAuth provider: %s", provider)
	}

	state := makeState()
	if err := s.rdb.Set(ctx, "nd:oauth:state:"+state, provider, 10*time.Minute).Err(); err != nil {
		return "", fmt.Errorf("store oauth state: %w", err)
	}

	redirectURI := s.cfg.OAuth2.RedirectBaseURL + "/api/v1/auth/oauth/" + provider + "/callback"

	params := url.Values{
		"client_id":     {prov.ClientID},
		"redirect_uri":  {redirectURI},
		"response_type": {"code"},
		"scope":         {prov.Scope},
		"state":         {state},
	}

	return prov.AuthURL + "?" + params.Encode(), nil
}

type OAuthCallbackResult struct {
	NewUser bool
	User    *UserResponse
	Tokens  *TokenPair
}

// OAuthCallback handles the OAuth callback after user authorization.
func (s *AuthService) OAuthCallback(ctx context.Context, provider, code, state string) (*OAuthCallbackResult, error) {
	// Verify state
	expected, err := s.rdb.Get(ctx, "nd:oauth:state:"+state).Result()
	if err != nil {
		return nil, model.ErrUnauthorized
	}
	if expected != provider {
		return nil, model.ErrUnauthorized
	}
	s.rdb.Del(ctx, "nd:oauth:state:"+state)

	prov, ok := s.cfg.OAuth2.Providers[provider]
	if !ok {
		return nil, fmt.Errorf("unknown OAuth provider: %s", provider)
	}

	redirectURI := s.cfg.OAuth2.RedirectBaseURL + "/api/v1/auth/oauth/" + provider + "/callback"

	// Exchange code for access token
	tokenResp, err := exchangeCode(prov, redirectURI, code)
	if err != nil {
		return nil, fmt.Errorf("exchange code: %w", err)
	}

	// Get user info
	userInfo, err := getUserInfo(prov, tokenResp.AccessToken)
	if err != nil {
		return nil, fmt.Errorf("get user info: %w", err)
	}

	// Find or create local user
	existingUser, err := s.queries.GetUserByEmail(ctx, oauthEmail(provider, userInfo.ID))
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return nil, fmt.Errorf("find user: %w", err)
	}

	var user sqlc.User
	isNew := false

	if err == nil {
		user = existingUser
	} else {
		isNew = true
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

		user, err = qtx.CreateUser(ctx, sqlc.CreateUserParams{
			Slug:         slug,
			Username:     userInfo.Username,
			Email:        oauthEmail(provider, userInfo.ID),
			PasswordHash: randomPasswordHash(),
		})
		if err != nil {
			if isUniqueViolation(err) {
				return nil, model.ErrAlreadyExists
			}
			return nil, fmt.Errorf("create user: %w", err)
		}

		displayName := userInfo.Username
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

		// Record the OAuth account binding
		_, err = s.queries.CreateOAuthAccount(ctx, sqlc.CreateOAuthAccountParams{
			UserID:            user.ID,
			Provider:          provider,
			ProviderAccountID: userInfo.ID,
			AccessToken:       tokenResp.AccessToken,
		})
		if err != nil {
			return nil, fmt.Errorf("create oauth account: %w", err)
		}
	}

	// Ensure OAuth account record exists for existing users
	if !isNew {
		_, err := s.queries.GetOAuthAccountByProviderAndID(ctx, sqlc.GetOAuthAccountByProviderAndIDParams{
			Provider:          provider,
			ProviderAccountID: userInfo.ID,
		})
		if err != nil {
			_, err = s.queries.CreateOAuthAccount(ctx, sqlc.CreateOAuthAccountParams{
				UserID:            user.ID,
				Provider:          provider,
				ProviderAccountID: userInfo.ID,
				AccessToken:       tokenResp.AccessToken,
			})
			if err != nil {
				return nil, fmt.Errorf("create oauth account for existing user: %w", err)
			}
		}
	}

	// Sync avatar from provider if available
	if userInfo.AvatarURL != "" {
		if err := s.queries.UpdateAvatarPath(ctx, sqlc.UpdateAvatarPathParams{
			UserID:     user.ID,
			AvatarPath: pgtype.Text{String: userInfo.AvatarURL, Valid: true},
		}); err != nil {
			return nil, fmt.Errorf("sync oauth avatar: %w", err)
		}
	}

	tokens, err := s.issueTokens(ctx, user.ID)
	if err != nil {
		return nil, err
	}

	resp := &UserResponse{
		Slug:           user.Slug,
		Username:       user.Username,
		Email:          user.Email,
		Status:         user.Status,
		Role:           user.Role,
		RegisterMethod: user.RegisterMethod,
		CreatedAt:      user.CreatedAt.Time.Format(time.RFC3339),
	}

	// Fetch profile
	profile, err := s.queries.GetProfileByUserID(ctx, user.ID)
	if err == nil && profile.ID != 0 {
		resp.Profile = ProfileData{
			DisplayName: profile.DisplayName.String,
			AvatarURL:   profile.AvatarPath.String,
			Bio:         profile.Bio.String,
		}
	} else {
		resp.Profile = ProfileData{DisplayName: user.Username}
	}

	// Fetch storage
	stats, err := s.queries.GetStorageStats(ctx, user.ID)
	if err == nil && stats.ID != 0 {
		resp.Storage = StorageData{
			StorageUsed:  stats.StorageUsed,
			StorageQuota: stats.StorageQuota,
		}
	}

	// Fetch level
	level, err := s.queries.GetLevelByUserID(ctx, user.ID)
	if err == nil && level.ID != 0 {
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

	return &OAuthCallbackResult{NewUser: isNew, User: resp, Tokens: tokens}, nil
}

func makeState() string {
	b := make([]byte, 32)
	rand.Read(b)
	return hex.EncodeToString(b)
}

func oauthEmail(provider, accountID string) string {
	return fmt.Sprintf("oauth_%s_%s@oauth.local", provider, accountID)
}

func randomPasswordHash() string {
	b := make([]byte, 32)
	rand.Read(b)
	hash, _ := bcrypt.GenerateFromPassword(b, bcrypt.DefaultCost)
	return string(hash)
}

func exchangeCode(prov config.OAuth2ProviderConfig, redirectURI, code string) (*OAuthTokenResponse, error) {
	form := url.Values{
		"grant_type":    {"authorization_code"},
		"code":          {code},
		"client_id":     {prov.ClientID},
		"client_secret": {prov.ClientSecret},
		"redirect_uri":  {redirectURI},
	}

	resp, err := http.PostForm(prov.TokenURL, form)
	if err != nil {
		return nil, fmt.Errorf("token request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read token response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("token endpoint returned %d: %s", resp.StatusCode, string(body))
	}

	// Try wrapper format first: { "c": 0, "d": { ... } }
	var wrapped struct {
		C int                 `json:"c"`
		D *OAuthTokenResponse `json:"d"`
	}
	if err := json.Unmarshal(body, &wrapped); err == nil && wrapped.D != nil {
		return wrapped.D, nil
	}

	// Fall back to flat format
	var token OAuthTokenResponse
	if err := json.Unmarshal(body, &token); err != nil {
		return nil, fmt.Errorf("parse token response: %w", err)
	}
	return &token, nil
}

func getUserInfo(prov config.OAuth2ProviderConfig, accessToken string) (*OAuthUserInfo, error) {
	req, err := http.NewRequest("GET", prov.UserInfoURL, nil)
	if err != nil {
		return nil, fmt.Errorf("create userinfo request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("userinfo request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read userinfo response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("userinfo endpoint returned %d: %s", resp.StatusCode, string(body))
	}

	// Try wrapper format first
	var wrapped struct {
		C int            `json:"c"`
		D *OAuthUserInfo `json:"d"`
	}
	if err := json.Unmarshal(body, &wrapped); err == nil && wrapped.D != nil {
		return wrapped.D, nil
	}

	// Fall back to flat format
	var info OAuthUserInfo
	if err := json.Unmarshal(body, &info); err != nil {
		return nil, fmt.Errorf("parse userinfo response: %w", err)
	}
	return &info, nil
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
