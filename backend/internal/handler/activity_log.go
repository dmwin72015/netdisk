package handler

import (
	"strconv"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/labstack/echo/v4"

	"github.com/netdisk/server/internal/db/sqlc"
	"github.com/netdisk/server/internal/middleware"
)

type ActivityLogHandler struct {
	queries *sqlc.Queries
}

func NewActivityLogHandler(queries *sqlc.Queries) *ActivityLogHandler {
	return &ActivityLogHandler{queries: queries}
}

type SecurityLogItem struct {
	ID        int64       `json:"id"`
	Action    string      `json:"action"`
	IP        pgtype.Text `json:"ip"`
	IPRegion  pgtype.Text `json:"ipRegion"`
	OS        pgtype.Text `json:"os"`
	Browser   pgtype.Text `json:"browser"`
	CreatedAt pgtype.Timestamptz `json:"createdAt"`
}

type LoginDevice struct {
	IP        string `json:"ip"`
	IPRegion  string `json:"ipRegion"`
	OS        string `json:"os"`
	Browser   string `json:"browser"`
	UserAgent string `json:"userAgent"`
	LastLogin string `json:"lastLogin"`
	IsCurrent bool   `json:"isCurrent"`
}

type SecurityLogListResponse struct {
	Items []SecurityLogItem `json:"items"`
	Total int64             `json:"total"`
}

// ListSecurityLogs returns paginated security-related activity logs.
func (h *ActivityLogHandler) ListSecurityLogs(c echo.Context) error {
	userID, _ := middleware.UserID(c)

	page, _ := strconv.Atoi(c.QueryParam("page"))
	if page < 1 {
		page = 1
	}
	pageSize, _ := strconv.Atoi(c.QueryParam("pageSize"))
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}
	offset := int32((page - 1) * pageSize)

	total, err := h.queries.CountSecurityLogsByUser(c.Request().Context(), userID)
	if err != nil {
		return err
	}

	rows, err := h.queries.ListSecurityLogsByUser(c.Request().Context(), sqlc.ListSecurityLogsByUserParams{
		UserID: userID,
		Limit:  int32(pageSize),
		Offset: offset,
	})
	if err != nil {
		return err
	}

	items := make([]SecurityLogItem, 0, len(rows))
	for _, r := range rows {
		items = append(items, SecurityLogItem{
			ID:        r.ID,
			Action:    r.Action,
			IP:        r.Ip,
			IPRegion:  r.IpRegion,
			OS:        r.Os,
			Browser:   r.Browser,
			CreatedAt: r.CreatedAt,
		})
	}

	return OK(c, SecurityLogListResponse{Items: items, Total: total})
}

// ListLoginDevices returns unique login devices based on activity logs.
func (h *ActivityLogHandler) ListLoginDevices(c echo.Context) error {
	userID, _ := middleware.UserID(c)

	// Get all login-related logs (up to 500 to find unique devices)
	rows, err := h.queries.ListSecurityLogsByUser(c.Request().Context(), sqlc.ListSecurityLogsByUserParams{
		UserID: userID,
		Limit:  500,
		Offset: 0,
	})
	if err != nil {
		return err
	}

	// Current session IP and the device id reported by the client (if any).
	currentIP := c.RealIP()
	currentDeviceID := c.QueryParam("deviceId")

	// Deduce unique devices by device id. Rows written without a device id
	// (older logs, or OAuth logins that fall back on the server) are grouped
	// by IP instead.
	type deviceKey struct {
		DeviceID string
	}
	seen := make(map[deviceKey]bool)
	var devices []LoginDevice

	for _, r := range rows {
		if r.Action != "user.login" && r.Action != "user.oauth_login" {
			continue
		}
		ip := r.Ip.String
		if ip == "" {
			continue
		}
		key := deviceKey{DeviceID: deviceKeyForRow(r, ip)}
		if seen[key] {
			continue
		}
		seen[key] = true

		isCurrent := false
		if currentDeviceID != "" && r.DeviceID.Valid {
			isCurrent = r.DeviceID.String == currentDeviceID
		} else {
			isCurrent = ip == currentIP
		}

		dev := LoginDevice{
			IP:        ip,
			IPRegion:  r.IpRegion.String,
			OS:        r.Os.String,
			Browser:   r.Browser.String,
			UserAgent: r.UserAgent.String,
			LastLogin: r.CreatedAt.Time.Format("2006-01-02 15:04:05"),
			IsCurrent: isCurrent,
		}
		devices = append(devices, dev)
	}

	if devices == nil {
		devices = []LoginDevice{}
	}

	return OK(c, devices)
}

// deviceKeyForRow returns the grouping key for a login log row: the stored
// device id when present, otherwise the IP (prefixed to avoid collisions with
// real device ids).
func deviceKeyForRow(r sqlc.UserActivityLog, ip string) string {
	if r.DeviceID.Valid && r.DeviceID.String != "" {
		return r.DeviceID.String
	}
	return "ip:" + ip
}
