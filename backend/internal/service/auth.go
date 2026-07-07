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
	queries   *sqlc.Queries
	pg        *pgxpool.Pool
	jwtMgr    *jwtutil.Manager
	cfg       *config.Config
	rdb       *redis.Client
	configSvc *SystemConfigService
}

func NewAuthService(queries *sqlc.Queries, pg *pgxpool.Pool, jwtMgr *jwtutil.Manager, cfg *config.Config, rdb *redis.Client, configSvc *SystemConfigService) *AuthService {
	return &AuthService{queries: queries, pg: pg, jwtMgr: jwtMgr, cfg: cfg, rdb: rdb, configSvc: configSvc}
}

func (s *AuthService) defaultQuota() int64 {
	if v, ok := s.configSvc.Get("default_quota"); ok {
		switch n := v.(type) {
		case int64:
			return n
		case float64:
			return int64(n)
		}
	}
	return s.cfg.Limits.DefaultStorageQuota
}

func (s *AuthService) Cfg() *config.Config {
	return s.cfg
}

type RegisterInput struct {
	Username string `json:"username"`
	Email    string `json:"email"`
	Password string `json:"password"`
	DeviceID string `json:"deviceId"`
}

type LoginInput struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	DeviceID string `json:"deviceId"`
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
	ID             int64       `json:"id"`
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
		Slug:           slug,
		Username:       input.Username,
		Email:          strText(input.Email),
		PasswordHash:   string(hash),
		RegisterMethod: "email",
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
		StorageQuota: s.defaultQuota(),
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
		ID:             user.ID,
		Slug:           user.Slug,
		Username:       user.Username,
		Email:          textStr(user.Email),
		Status:         user.Status,
		Role:           user.Role,
		RegisterMethod: user.RegisterMethod,
		Profile: ProfileData{
			DisplayName: user.Username,
		},
		Storage: StorageData{
			StorageQuota: s.defaultQuota(),
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

	user, err := s.queries.GetUserByEmail(ctx, strText(input.Email))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil, model.ErrInvalidCredentials
		}
		return nil, nil, fmt.Errorf("get user: %w", err)
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(input.Password)); err != nil {
		return nil, nil, model.ErrInvalidCredentials
	}

	tokens, err := s.issueTokens(ctx, user.ID)
	if err != nil {
		return nil, nil, err
	}

	resp := &UserResponse{
		ID:             user.ID,
		Slug:           user.Slug,
		Username:       user.Username,
		Email:          textStr(user.Email),
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
	UserID  int64
	User    *UserResponse
	Tokens  *TokenPair
	// Bind flow fields
	IsBind               bool   // true when this is a bind (not login) result
	Provider             string // the provider that was bound
	AlreadyBound         bool   `json:"alreadyBound"`         // true when bind flow detected the account is already bound to the current user
	NeedReplaceConfirm   bool   `json:"needReplaceConfirm"`   // true when the current user already has a binding for the same provider with a different third-party account
	ReplaceToken         string `json:"replaceToken"`         // one-time token to confirm the replacement
	ReplaceProvider      string `json:"replaceProvider"`      // the provider that would be replaced (mirrors Provider for the frontend)
	OldProviderAccountID string `json:"oldProviderAccountId"` // the existing provider_account_id that would be replaced (helps UI identify it)
	OldOAuthEmail        string `json:"oldOauthEmail"`        // the existing oauth_email that would be replaced (helps UI identify it)
	// Email match confirm flow (login path)
	EmailMatchConfirm bool   `json:"emailMatchConfirm"` // true when the third-party email matches an existing user, needs confirmation
	ConfirmToken      string `json:"confirmToken"`      // one-time token to complete the email-match + linking
	MatchEmail        string `json:"matchEmail"`        // the matched email for display
	MatchProvider     string `json:"matchProvider"`     // the OAuth provider name for display
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
		// 1. Check if the third-party account is already bound somewhere
		existing, err := s.queries.GetOAuthAccountByProviderAndID(ctx, sqlc.GetOAuthAccountByProviderAndIDParams{
			Provider:          provider,
			ProviderAccountID: userInfo.ID,
		})
		if err == nil {
			if existing.UserID != bindUserID {
				// Already bound to another user — reject. The user must unbind
				// it from the original account before binding it here.
				errMsg := fmt.Sprintf("此 %s 账号已被其他账号关联，请先在原账号解绑后再来关联", provider)
				return &OAuthCallbackResult{IsBind: true, Provider: provider}, errors.New(errMsg)
			}
			// Already bound to the current user — no-op. The user-facing
			// email is left untouched; we do not refresh access_token / oauth_email
			// here per requirements.
			return &OAuthCallbackResult{IsBind: true, Provider: provider, AlreadyBound: true}, nil
		} else if !errors.Is(err, pgx.ErrNoRows) {
			return &OAuthCallbackResult{IsBind: true, Provider: provider}, fmt.Errorf("find oauth account: %w", err)
		}

		// 2. Third-party account is free. Check whether the current user already
		// has a binding for the same provider (a different third-party account).
		// If so, we need explicit user confirmation before replacing it.
		oldProviderAccountID, oldOAuthEmail, hasOldForProvider, err := s.findUserOAuthForProvider(ctx, bindUserID, provider)
		if err != nil {
			return &OAuthCallbackResult{IsBind: true, Provider: provider}, err
		}

		if hasOldForProvider {
			// Mint a one-time replace token (10 min) and ask the user to confirm.
			replaceToken := makeState()
			replaceVal := fmt.Sprintf("%s:%d:%s:%s:%s", provider, bindUserID, userInfo.ID, userInfo.Email, tokenResp.AccessToken)
			if err := s.rdb.Set(ctx, "nd:oauth:bind-replace:"+replaceToken, replaceVal, 10*time.Minute).Err(); err != nil {
				return &OAuthCallbackResult{IsBind: true, Provider: provider}, fmt.Errorf("store replace token: %w", err)
			}
			return &OAuthCallbackResult{
				IsBind:               true,
				Provider:             provider,
				NeedReplaceConfirm:   true,
				ReplaceToken:         replaceToken,
				ReplaceProvider:      provider,
				OldProviderAccountID: oldProviderAccountID,
				OldOAuthEmail:        oldOAuthEmail,
			}, nil
		}

		// 3. No prior binding for this provider — bind directly.
		if err := s.upsertOAuthAccount(ctx, bindUserID, provider, userInfo.ID, userInfo.Email, tokenResp.AccessToken); err != nil {
			return &OAuthCallbackResult{IsBind: true, Provider: provider}, fmt.Errorf("upsert oauth account: %w", err)
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

	// Login path: look up an existing binding for the third-party account.
	// If none, auto-create a new local user and bind the OAuth account.
	user, err := s.resolveOAuthUser(ctx, provider, userInfo)
	if err != nil {
		return nil, err
	}
	if user == nil {
		// No prior binding — check whether the email matches an existing
		// user (e.g. after unbind and re-login). If so, require explicit
		// confirmation from the user before linking and logging in.
		if userInfo.Email != "" {
			existingUser, err := s.queries.GetUserByEmail(ctx, strText(userInfo.Email))
			if err == nil {
				confirmToken := makeState()
				confirmVal := fmt.Sprintf("%s:%d:%s:%s:%s", provider, existingUser.ID, userInfo.ID, userInfo.Email, tokenResp.AccessToken)
				if err := s.rdb.Set(ctx, "nd:oauth:email-confirm:"+confirmToken, confirmVal, 10*time.Minute).Err(); err != nil {
					return nil, fmt.Errorf("store email confirm token: %w", err)
				}
				return &OAuthCallbackResult{
					EmailMatchConfirm: true,
					ConfirmToken:      confirmToken,
					MatchEmail:        userInfo.Email,
					MatchProvider:     provider,
				}, nil
			}
			if !errors.Is(err, pgx.ErrNoRows) {
				return nil, fmt.Errorf("find user by email: %w", err)
			}
		}

		// No match — auto-register a new local user.
		newUser, err := s.createUserFromOAuth(ctx, provider, userInfo)
		if err != nil {
			return nil, err
		}
		user = newUser
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
		ID:             user.ID,
		Slug:           user.Slug,
		Username:       user.Username,
		Email:          textStr(user.Email),
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

	return &OAuthCallbackResult{UserID: user.ID, User: resp, Tokens: tokens, NewUser: true}, nil
}

func (s *AuthService) upsertOAuthAccount(ctx context.Context, userID int64, provider, providerAccountID, oauthEmail, accessToken string) error {
	_, err := s.pg.Exec(ctx, `
		INSERT INTO user_oauth_accounts (user_id, provider, provider_account_id, access_token, oauth_email, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, NOW(), NOW())
		ON CONFLICT (provider, provider_account_id)
		DO UPDATE SET
			access_token = EXCLUDED.access_token,
			oauth_email  = COALESCE(EXCLUDED.oauth_email, user_oauth_accounts.oauth_email),
			updated_at   = NOW()
	`, userID, provider, providerAccountID, accessToken, nullIfEmpty(oauthEmail))
	if err != nil {
		return fmt.Errorf("upsert oauth account: %w", err)
	}
	return nil
}

// findUserOAuthForProvider returns the existing (provider_account_id, oauth_email)
// for the given user+provider, or ("", "", false, nil) if there is no binding.
func (s *AuthService) findUserOAuthForProvider(ctx context.Context, userID int64, provider string) (string, string, bool, error) {
	var providerAccountID, oauthEmail string
	err := s.pg.QueryRow(ctx, `
		SELECT provider_account_id, COALESCE(oauth_email, '')
		FROM user_oauth_accounts
		WHERE user_id = $1 AND provider = $2
		LIMIT 1
	`, userID, provider).Scan(&providerAccountID, &oauthEmail)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return "", "", false, nil
		}
		return "", "", false, fmt.Errorf("find user oauth: %w", err)
	}
	return providerAccountID, oauthEmail, true, nil
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
	if !user.Email.Valid && len(accounts) == 1 {
		return model.ErrForbidden
	}

	_, err = s.pg.Exec(ctx, `DELETE FROM user_oauth_accounts WHERE user_id = $1 AND provider = $2`, userID, provider)
	if err != nil {
		return fmt.Errorf("delete oauth account: %w", err)
	}
	return nil
}

// ConfirmOAuthBindReplace consumes a one-time replace token and replaces the
// current user's existing binding for the same provider with the third-party
// account embedded in the token. It performs the swap atomically and deletes
// the token to prevent replay.
func (s *AuthService) ConfirmOAuthBindReplace(ctx context.Context, replaceToken string, userID int64) (string, error) {
	if replaceToken == "" {
		return "", model.ErrInvalidInput
	}

	// Retrieve and delete the token immediately to prevent replay.
	replaceVal, err := s.rdb.GetDel(ctx, "nd:oauth:bind-replace:"+replaceToken).Result()
	if err != nil {
		return "", model.ErrUnauthorized
	}

	// Parse stored data: provider:bindUserID:providerAccountID:email:accessToken
	parts := strings.SplitN(replaceVal, ":", 5)
	if len(parts) != 5 {
		return "", model.ErrUnauthorized
	}
	provider := parts[0]
	storedBindUserID, err := strconv.ParseInt(parts[1], 10, 64)
	if err != nil {
		return "", model.ErrUnauthorized
	}
	providerAccountID := parts[2]
	email := parts[3]
	accessToken := parts[4]

	if storedBindUserID != userID {
		return "", model.ErrForbidden
	}

	tx, err := s.pg.Begin(ctx)
	if err != nil {
		return "", fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback(ctx)

	// Drop any existing bindings the current user has for this provider.
	// The new binding has not been written yet, so this only removes the
	// pre-existing record (its provider_account_id will differ).
	if _, err := tx.Exec(ctx,
		`DELETE FROM user_oauth_accounts WHERE user_id = $1 AND provider = $2`,
		userID, provider,
	); err != nil {
		return "", fmt.Errorf("delete existing provider bindings: %w", err)
	}

	// Insert the new binding inside the same transaction.
	if _, err := tx.Exec(ctx, `
		INSERT INTO user_oauth_accounts (user_id, provider, provider_account_id, access_token, oauth_email, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, NOW(), NOW())
		ON CONFLICT (provider, provider_account_id)
		DO UPDATE SET
			access_token = EXCLUDED.access_token,
			oauth_email  = COALESCE(EXCLUDED.oauth_email, user_oauth_accounts.oauth_email),
			updated_at   = NOW()
	`, userID, provider, providerAccountID, accessToken, nullIfEmpty(email)); err != nil {
		return "", fmt.Errorf("upsert oauth account: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return "", fmt.Errorf("commit: %w", err)
	}

	// Best-effort avatar sync (outside the tx). Only fills in an avatar
	// when the user does not already have one — never overwrites a
	// user-set avatar.
	if prov, ok := s.cfg.OAuth2.Providers[provider]; ok {
		if info, err := getUserInfo(prov, accessToken, provider); err == nil && info.AvatarURL != "" {
			existingProfile, _ := s.queries.GetProfileByUserID(ctx, userID)
			if existingProfile.ID == 0 || existingProfile.AvatarPath.String == "" {
				_ = s.queries.UpdateAvatarPath(ctx, sqlc.UpdateAvatarPathParams{
					UserID:     userID,
					AvatarPath: pgtype.Text{String: info.AvatarURL, Valid: true},
				})
			}
		}
	}

	return provider, nil
}

// ConfirmEmailOAuthLink consumes a one-time confirm token from the login
// email-match flow. It links the OAuth account to the existing user and
// issues tokens atomically, then returns the token pair.
func (s *AuthService) ConfirmEmailOAuthLink(ctx context.Context, confirmToken string) (*TokenPair, error) {
	if confirmToken == "" {
		return nil, model.ErrInvalidInput
	}

	confirmVal, err := s.rdb.GetDel(ctx, "nd:oauth:email-confirm:"+confirmToken).Result()
	if err != nil {
		return nil, model.ErrUnauthorized
	}

	// Parse stored data: provider:userID:providerAccountID:email:accessToken
	parts := strings.SplitN(confirmVal, ":", 5)
	if len(parts) != 5 {
		return nil, model.ErrUnauthorized
	}
	provider := parts[0]
	userID, err := strconv.ParseInt(parts[1], 10, 64)
	if err != nil {
		return nil, model.ErrUnauthorized
	}
	providerAccountID := parts[2]
	email := parts[3]
	accessToken := parts[4]

	if err := s.upsertOAuthAccount(ctx, userID, provider, providerAccountID, email, accessToken); err != nil {
		return nil, fmt.Errorf("upsert oauth account: %w", err)
	}

	tokens, err := s.issueTokens(ctx, userID)
	if err != nil {
		return nil, err
	}

	return tokens, nil
}

func (s *AuthService) resolveOAuthUser(ctx context.Context, provider string, userInfo *OAuthUserInfo) (*sqlc.User, error) {
	// Look up an existing binding for this third-party account.
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

	// No binding found. The caller (OAuthCallback) will auto-create a
	// new local user. Email-based matching is intentionally not performed.
	return nil, nil
}

// createUserFromOAuth registers a brand-new local user derived from the
// third-party profile, and persists the corresponding register_method.
func (s *AuthService) createUserFromOAuth(ctx context.Context, provider string, userInfo *OAuthUserInfo) (*sqlc.User, error) {
	slug, err := gonanoid.New(21)
	if err != nil {
		return nil, fmt.Errorf("generate slug: %w", err)
	}

	username := userInfo.Username
	if username == "" {
		// Fall back to a slug-based username when the provider omitted one.
		username = "user_" + slug[:8]
	}

	var emailArg string
	registerMethod := provider
	if userInfo.Email != "" {
		emailArg = userInfo.Email
	}

	tx, err := s.pg.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback(ctx)

	var user sqlc.User
	if err := tx.QueryRow(ctx, `
		INSERT INTO users (slug, username, email, password_hash, register_method)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, slug, username, email, password_hash, status, created_at, updated_at, register_method, role
	`, slug, username, emailArg, randomPasswordHash(), registerMethod).Scan(
		&user.ID,
		&user.Slug,
		&user.Username,
		&user.Email,
		&user.PasswordHash,
		&user.Status,
		&user.CreatedAt,
		&user.UpdatedAt,
		&user.RegisterMethod,
		&user.Role,
	); err != nil {
		if isUniqueViolation(err) {
			return nil, model.ErrAlreadyExists
		}
		return nil, fmt.Errorf("create oauth user: %w", err)
	}

	// Create the companion profile, storage stats, and level rows so the
	// new user is functionally identical to one created via /register.
	if _, err := tx.Exec(ctx, `
		INSERT INTO user_profiles (user_id, display_name) VALUES ($1, $2)
	`, user.ID, username); err != nil {
		return nil, fmt.Errorf("create oauth user profile: %w", err)
	}

	if _, err := tx.Exec(ctx, `
		INSERT INTO user_storage_stats (user_id, storage_quota) VALUES ($1, $2)
	`, user.ID, s.defaultQuota()); err != nil {
		return nil, fmt.Errorf("create oauth storage stats: %w", err)
	}

	if _, err := tx.Exec(ctx, `
		INSERT INTO user_levels (user_id, level_code, level_name) VALUES ($1, $2, $3)
	`, user.ID, "vip1", "VIP1"); err != nil {
		return nil, fmt.Errorf("create oauth user level: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("commit: %w", err)
	}

	return &user, nil
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
