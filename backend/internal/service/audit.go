package service

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"strings"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/zerolog"

	"github.com/netdisk/server/internal/db/sqlc"
	"github.com/netdisk/server/internal/pkg/iplookup"
	"github.com/netdisk/server/internal/pkg/useragent"
)

// Audit action constants.
const (
	ActionLogin           = "user.login"
	ActionRegister        = "user.register"
	ActionLogout          = "user.logout"
	ActionOAuthLogin      = "user.oauth_login"
	ActionPasswordChange  = "user.password_change"
	ActionFileUpload      = "file.upload"
	ActionFileDelete      = "file.delete"
	ActionFileRename      = "file.rename"
	ActionFileMove        = "file.move"
	ActionDirCreate       = "dir.create"
	ActionDirLock         = "dir.lock"
	ActionDirUnlock       = "dir.unlock"
	ActionDirUnlockFailed = "dir.unlock_failed"
	ActionShareCreate     = "share.create"
	ActionShareDelete     = "share.delete"
	ActionAdminCreateUser = "admin.create_user"
	ActionAdminDeleteUser = "admin.delete_user"
	ActionAdminUpdateRole = "admin.update_role"
)

type AuditService struct {
	pg     *pgxpool.Pool
	lookup iplookup.Lookup
	logger zerolog.Logger
}

func NewAuditService(pg *pgxpool.Pool, lookup iplookup.Lookup, logger zerolog.Logger) *AuditService {
	return &AuditService{
		pg:     pg,
		lookup: lookup,
		logger: logger.With().Str("component", "audit").Logger(),
	}
}

type AuditLogInput struct {
	UserID       int64
	Action       string
	ResourceType string
	ResourceName string
	IP           string
	UserAgent    string
	DeviceID     string
	Extra        map[string]any
}

// Log writes an activity log entry asynchronously.
func (s *AuditService) Log(ctx context.Context, input AuditLogInput) {
	go s.write(input)
}

func (s *AuditService) write(input AuditLogInput) {
	uaInfo := useragent.Parse(input.UserAgent)

	var region iplookup.Region
	if s.lookup != nil {
		var err error
		region, err = s.lookup.Lookup(input.IP)
		if err != nil {
			s.logger.Warn().Err(err).Str("ip", input.IP).Msg("ip lookup failed")
		}
	}

	var extra []byte
	if len(input.Extra) > 0 {
		extra, _ = json.Marshal(input.Extra)
	}

	// Use the client-supplied device id, or derive a stable fallback from the
	// request IP and User-Agent when none was provided (e.g. OAuth logins).
	deviceID := input.DeviceID
	if deviceID == "" {
		deviceID = fallbackDeviceID(input.IP, input.UserAgent)
	}

	queries := sqlc.New(s.pg)
	err := queries.CreateActivityLog(context.Background(), sqlc.CreateActivityLogParams{
		UserID:       input.UserID,
		Action:       input.Action,
		ResourceType: toText(input.ResourceType),
		ResourceName: toText(input.ResourceName),
		Ip:           toText(input.IP),
		IpRegion:     toText(region.String()),
		UserAgent:    toText(input.UserAgent),
		Os:           toText(uaInfo.OS),
		Browser:      toText(uaInfo.Browser),
		DeviceID:     toText(deviceID),
		Extra:        extra,
	})
	if err != nil {
		s.logger.Warn().Err(err).Str("action", input.Action).Msg("failed to write activity log")
	}
}

func toText(s string) pgtype.Text {
	if s == "" {
		return pgtype.Text{}
	}
	return pgtype.Text{String: s, Valid: true}
}

// fallbackDeviceID derives a stable device identifier from the request IP and a
// normalized User-Agent when the client did not supply one (e.g. OAuth logins,
// where a device id cannot be forwarded through the redirect flow). The "fb-"
// prefix avoids collisions with real client-supplied device ids.
func fallbackDeviceID(ip, ua string) string {
	normalized := strings.ToLower(strings.TrimSpace(ua))
	sum := sha256.Sum256([]byte(ip + "|" + normalized))
	return "fb-" + hex.EncodeToString(sum[:])
}
