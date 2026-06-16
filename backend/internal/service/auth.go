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
	"strconv"
	"strings"
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
	Email     string `json:"email"`
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
	// Bind flow fields
	IsBind        bool   // true when this is a bind (not login) result
	Provider      string // the provider that was bound
	Conflict      bool   // true when the OAuth account is already bound to another user
	ConflictToken string // token to confirm the rebind
}

// OAuthAuthorizeForBind returns the authorization URL for binding an OAuth account to the current user.
func (s *AuthService) OAuthAuthorizeForBind(ctx context.Context, provider string, userID int64) (string, error) {
	prov, ok := s.cfg.OAuth2.Providers[provider]
	if !ok {
		return "", fmt.Errorf("unknown OAuth provider: %s", provider)
	}

	state := makeState()
	// Store state with bind mode and user ID
	stateValue := fmt.Sprintf("%s:bind:%d", provider, userID)
	if err := s.rdb.Set(ctx, "nd:oauth:state:"+state, stateValue, 10*time.Minute).Err(); err != nil {
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

// OAuthCallback handles the OAuth callback after user authorization.
func (s *AuthService) OAuthCallback(ctx context.Context, provider, code, state string) (*OAuthCallbackResult, error) {
	// Verify state
	expected, err := s.rdb.Get(ctx, "nd:oauth:state:"+state).Result()
	if err != nil {
		return nil, model.ErrUnauthorized
	}
	s.rdb.Del(ctx, "nd:oauth:state:"+state)

	// Check if this is a bind flow
	var bindUserID int64
	parts := strings.SplitN(expected, ":", 3)
	if len(parts) == 3 && parts[1] == "bind" {
		if parts[0] != provider {
			return nil, model.ErrUnauthorized
		}
		bindUserID, err = strconv.ParseInt(parts[2], 10, 64)
		if err != nil {
			return nil, model.ErrUnauthorized
		}
	} else if expected != provider {
		return nil, model.ErrUnauthorized
	}

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
	userInfo, err := getUserInfo(prov, tokenResp.AccessToken, provider)
	if err != nil {
		return nil, fmt.Errorf("get user info: %w", err)
	}

	// Bind flow: link OAuth account to existing user without creating user or issuing tokens
	if bindUserID != 0 {
		// Check if OAuth account is already bound to another user
		existing, err := s.queries.GetOAuthAccountByProviderAndID(ctx, sqlc.GetOAuthAccountByProviderAndIDParams{
			Provider:          provider,
			ProviderAccountID: userInfo.ID,
		})
		if err == nil {
			if existing.UserID != bindUserID {
				// Conflict: already bound to another user — store confirm token
				confirmToken := makeState()
				confirmVal := fmt.Sprintf("%s:%d:%s:%s:%s", provider, bindUserID, userInfo.ID, userInfo.Email, tokenResp.AccessToken)
				if err := s.rdb.Set(ctx, "nd:oauth:confirm:"+confirmToken, confirmVal, 10*time.Minute).Err(); err != nil {
					return nil, fmt.Errorf("store confirm token: %w", err)
				}
				return &OAuthCallbackResult{
					IsBind:        true,
					Provider:      provider,
					Conflict:      true,
					ConflictToken: confirmToken,
				}, nil
			}
			return &OAuthCallbackResult{IsBind: true, Provider: provider}, errors.New("this " + provider + " account is already linked to your account")
		} else if !errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("find oauth account: %w", err)
		}

		// Not bound yet — insert
		if err := s.upsertOAuthAccount(ctx, bindUserID, provider, userInfo.ID, userInfo.Email, tokenResp.AccessToken); err != nil {
			return nil, fmt.Errorf("upsert oauth account: %w", err)
		}

		// Sync avatar if user doesn't have one yet
		if userInfo.AvatarURL != "" {
			existingProfile, _ := s.queries.GetProfileByUserID(ctx, bindUserID)
			if existingProfile.ID == 0 || existingProfile.AvatarPath.String == "" {
				_ = s.queries.UpdateAvatarPath(ctx, sqlc.UpdateAvatarPathParams{
					UserID:     bindUserID,
					AvatarPath: pgtype.Text{String: userInfo.AvatarURL, Valid: true},
				})
			}
		}

		return &OAuthCallbackResult{IsBind: true, Provider: provider}, nil
	}

	// Find or create local user by OAuth binding first, then by email
	user, err := s.resolveOAuthUser(ctx, provider, userInfo)
	if err != nil {
		return nil, err
	}
	isNew := user == nil

	if isNew {
		slug, err := gonanoid.New(21)
		if err != nil {
			return nil, fmt.Errorf("generate slug: %w", err)
		}

		tx, err := s.pg.Begin(ctx)
		if err != nil {
			return nil, fmt.Errorf("begin tx: %w", err)
		}
		defer tx.Rollback(ctx)

		// Use raw SQL to support NULL email
		var registerMethod string
		if userInfo.Email != "" {
			registerMethod = "oauth"
		} else {
			registerMethod = "oauth_noemail"
		}
		var newUser sqlc.User
		err = tx.QueryRow(ctx, `
			INSERT INTO users (slug, username, email, password_hash, register_method)
			VALUES ($1, $2, $3, $4, $5)
			RETURNING id, slug, username, email, password_hash, status, created_at, updated_at, register_method, role
		`, slug, userInfo.Username, nullIfEmpty(userInfo.Email), randomPasswordHash(), registerMethod).Scan(
			&newUser.ID, &newUser.Slug, &newUser.Username, &newUser.Email, &newUser.PasswordHash,
			&newUser.Status, &newUser.CreatedAt, &newUser.UpdatedAt, &newUser.RegisterMethod, &newUser.Role,
		)
		if err != nil {
			if isUniqueViolation(err) {
				return nil, model.ErrAlreadyExists
			}
			return nil, fmt.Errorf("create user: %w", err)
		}
		user = &newUser

		displayName := userInfo.Username
		_, err = tx.Exec(ctx, `
			INSERT INTO user_profiles (user_id, display_name)
			VALUES ($1, $2)
		`, user.ID, displayName)
		if err != nil {
			return nil, fmt.Errorf("create profile: %w", err)
		}

		_, err = tx.Exec(ctx, `
			INSERT INTO user_storage_stats (user_id, storage_used, storage_quota)
			VALUES ($1, 0, $2)
		`, user.ID, s.cfg.Limits.DefaultStorageQuota)
		if err != nil {
			return nil, fmt.Errorf("create storage stats: %w", err)
		}

		_, err = tx.Exec(ctx, `
			INSERT INTO user_levels (user_id, level_code, level_name)
			VALUES ($1, 'vip1', 'VIP1')
		`, user.ID)
		if err != nil {
			return nil, fmt.Errorf("create level: %w", err)
		}

		if err := tx.Commit(ctx); err != nil {
			return nil, fmt.Errorf("commit: %w", err)
		}
	}

	if err := s.upsertOAuthAccount(ctx, user.ID, provider, userInfo.ID, userInfo.Email, tokenResp.AccessToken); err != nil {
		return nil, fmt.Errorf("upsert oauth account: %w", err)
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

func (s *AuthService) upsertOAuthAccount(ctx context.Context, userID int64, provider, providerAccountID, oauthEmail, accessToken string) error {
	if oauthEmail != "" {
		_, err := s.pg.Exec(ctx, `
			INSERT INTO user_oauth_accounts (user_id, provider, provider_account_id, access_token, oauth_email, created_at, updated_at)
			VALUES ($1, $2, $3, $4, $5, NOW(), NOW())
			ON CONFLICT (provider, provider_account_id)
			DO UPDATE SET access_token = EXCLUDED.access_token, updated_at = NOW()
		`, userID, provider, providerAccountID, accessToken, oauthEmail)
		return err
	}
	_, err := s.pg.Exec(ctx, `
		INSERT INTO user_oauth_accounts (user_id, provider, provider_account_id, access_token, created_at, updated_at)
		VALUES ($1, $2, $3, $4, NOW(), NOW())
		ON CONFLICT (provider, provider_account_id)
		DO UPDATE SET access_token = EXCLUDED.access_token, updated_at = NOW()
	`, userID, provider, providerAccountID, accessToken)
	return err
}

func (s *AuthService) OAuthUnlink(ctx context.Context, userID int64, provider string) error {
	user, err := s.queries.GetUserByID(ctx, userID)
	if err != nil {
		return fmt.Errorf("get user: %w", err)
	}

	accounts, err := s.queries.GetOAuthAccountsByUserID(ctx, userID)
	if err != nil {
		return fmt.Errorf("get oauth accounts: %w", err)
	}

	var found bool
	for _, a := range accounts {
		if a.Provider == provider {
			found = true
			break
		}
	}
	if !found {
		return model.ErrNotFound
	}

	// If user has no email and this is the only OAuth login method, block unlinking
	if user.Email == "" && len(accounts) == 1 {
		return model.ErrForbidden
	}

	_, err = s.pg.Exec(ctx, `DELETE FROM user_oauth_accounts WHERE user_id = $1 AND provider = $2`, userID, provider)
	if err != nil {
		return fmt.Errorf("delete oauth account: %w", err)
	}
	return nil
}

func (s *AuthService) resolveOAuthUser(ctx context.Context, provider string, userInfo *OAuthUserInfo) (*sqlc.User, error) {
	// 1. Look up by OAuth account binding
	oa, err := s.queries.GetOAuthAccountByProviderAndID(ctx, sqlc.GetOAuthAccountByProviderAndIDParams{
		Provider:          provider,
		ProviderAccountID: userInfo.ID,
	})
	if err == nil {
		u, err := s.queries.GetUserByID(ctx, oa.UserID)
		if err != nil {
			return nil, fmt.Errorf("get oauth user: %w", err)
		}
		return &u, nil
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		return nil, fmt.Errorf("find oauth account: %w", err)
	}

	// 2. Look up by real email (link existing email-registered user)
	if userInfo.Email != "" {
		existing, err := s.queries.GetUserByEmail(ctx, userInfo.Email)
		if err == nil {
			return &existing, nil
		}
		if !errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("find user by email: %w", err)
		}
	}

	// 3. New user
	return nil, nil
}

func makeState() string {
	b := make([]byte, 32)
	rand.Read(b)
	return hex.EncodeToString(b)
}

func nullIfEmpty(s string) *string {
	if s == "" {
		return nil
	}
	return &s
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

	req, err := http.NewRequest("POST", prov.TokenURL, strings.NewReader(form.Encode()))
	if err != nil {
		return nil, fmt.Errorf("create token request: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "application/json")

	resp, err := http.DefaultClient.Do(req)
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

func getUserInfo(prov config.OAuth2ProviderConfig, accessToken, provider string) (*OAuthUserInfo, error) {
	req, err := http.NewRequest("GET", prov.UserInfoURL, nil)
	if err != nil {
		return nil, fmt.Errorf("create userinfo request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Accept", "application/json")

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

	// Provider-specific parsing
	switch provider {
	case "github":
		var gh struct {
			ID        int64  `json:"id"`
			Login     string `json:"login"`
			AvatarURL string `json:"avatar_url"`
		}
		if err := json.Unmarshal(body, &gh); err != nil {
			return nil, fmt.Errorf("parse github userinfo: %w", err)
		}
		info := &OAuthUserInfo{
			ID:        strconv.FormatInt(gh.ID, 10),
			Username:  gh.Login,
			AvatarURL: gh.AvatarURL,
		}
		// Fetch primary verified email from /user/emails
		info.Email = fetchGitHubPrimaryEmail(accessToken)
		return info, nil
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

func fetchGitHubPrimaryEmail(accessToken string) string {
	req, err := http.NewRequest("GET", "https://api.github.com/user/emails", nil)
	if err != nil {
		return ""
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Accept", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return ""
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return ""
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return ""
	}

	var emails []struct {
		Email    string `json:"email"`
		Primary  bool   `json:"primary"`
		Verified bool   `json:"verified"`
	}
	if err := json.Unmarshal(body, &emails); err != nil {
		return ""
	}

	for _, e := range emails {
		if e.Primary && e.Verified {
			return e.Email
		}
	}
	// Fall back to any verified email
	for _, e := range emails {
		if e.Verified {
			return e.Email
		}
	}
	return ""
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
