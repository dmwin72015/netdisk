package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/labstack/echo/v4"

	"github.com/netdisk/server/internal/model"
	"github.com/netdisk/server/internal/service"
)

type ShareHandler struct {
	svc   *service.ShareService
	audit *service.AuditService
}

func NewShareHandler(svc *service.ShareService, audit *service.AuditService) *ShareHandler {
	return &ShareHandler{svc: svc, audit: audit}
}

type shareRequest struct {
	FileSlugs    []string `json:"fileSlugs"`
	PasswordCode *string  `json:"passwordCode"`
	ExpiresAt    *string  `json:"expiresAt"`
}

type createShareRequest struct {
	FileSlugs    []string `json:"fileSlugs"`
	PasswordCode *string  `json:"passwordCode"`
	ExpiresAt    *string  `json:"expiresAt"`
}

type updateShareRequest struct {
	PasswordCode    *string `json:"passwordCode"`
	PasswordCodeSet bool    `json:"-"`
	ExpiresAt       *string `json:"expiresAt"`
	ExpiresAtSet    bool    `json:"-"`
}

type verifyShareRequest struct {
	PasswordCode string `json:"passwordCode"`
}

func (h *ShareHandler) CreateShare(c echo.Context) error {
	userID, err := requireUserID(c)
	if err != nil {
		return err
	}

	var request createShareRequest
	if err := c.Bind(&request); err != nil {
		return BizError(model.ErrInvalidInput)
	}
	expiresAt, err := parseOptionalShareTime(request.ExpiresAt)
	if err != nil {
		return err
	}

	share, err := h.svc.CreateShare(c.Request().Context(), userID, service.CreateShareInput{
		FileSlugs:    request.FileSlugs,
		PasswordCode: request.PasswordCode,
		ExpiresAt:    expiresAt,
	})
	if err != nil {
		return BizError(err)
	}
	h.audit.Log(c.Request().Context(), service.AuditLogInput{
		UserID: userID, Action: service.ActionShareCreate,
		ResourceType: "share",
		IP: c.RealIP(), UserAgent: c.Request().UserAgent(),
		Extra: map[string]any{"fileCount": len(request.FileSlugs)},
	})
	return Created(c, share)
}

func (h *ShareHandler) ListShares(c echo.Context) error {
	userID, err := requireUserID(c)
	if err != nil {
		return err
	}

	page, _ := strconv.Atoi(c.QueryParam("page"))
	pageSize, _ := strconv.Atoi(c.QueryParam("pageSize"))
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 || pageSize > 100 {
		pageSize = 50
	}

	shares, total, err := h.svc.ListShares(c.Request().Context(), userID, pageSize, (page-1)*pageSize)
	if err != nil {
		return BizError(err)
	}
	return OK(c, map[string]any{"shares": shares, "total": total})
}

func (h *ShareHandler) UpdateShare(c echo.Context) error {
	userID, err := requireUserID(c)
	if err != nil {
		return err
	}

	var request updateShareRequest
	if err := c.Bind(&request); err != nil {
		return BizError(model.ErrInvalidInput)
	}
	expiresAt, err := parseOptionalShareTime(request.ExpiresAt)
	if err != nil {
		return err
	}

	share, err := h.svc.UpdateShare(c.Request().Context(), userID, c.Param("slug"), service.UpdateShareInput{
		PasswordCode:    request.PasswordCode,
		PasswordCodeSet: request.PasswordCodeSet,
		ExpiresAt:       expiresAt,
		ExpiresAtSet:    request.ExpiresAtSet,
	})
	if err != nil {
		return BizError(err)
	}
	return OK(c, share)
}

func (h *ShareHandler) CancelShare(c echo.Context) error {
	userID, err := requireUserID(c)
	if err != nil {
		return err
	}

	slug := c.Param("slug")
	if c.QueryParam("permanent") == "1" {
		if err := h.svc.DeleteShare(c.Request().Context(), userID, slug); err != nil {
			return BizError(err)
		}
		h.audit.Log(c.Request().Context(), service.AuditLogInput{
			UserID: userID, Action: service.ActionShareDelete,
			ResourceType: "share", ResourceName: slug,
			IP: c.RealIP(), UserAgent: c.Request().UserAgent(),
		})
		return OK(c, map[string]string{"message": "share deleted"})
	}

	if err := h.svc.CancelShare(c.Request().Context(), userID, slug); err != nil {
		return BizError(err)
	}
	return OK(c, map[string]string{"message": "share canceled"})
}

func (h *ShareHandler) GetPublicShare(c echo.Context) error {
	info, err := h.svc.GetPublicInfo(c.Request().Context(), c.Param("slug"))
	if err != nil {
		return BizError(err)
	}
	return OK(c, info)
}

func (h *ShareHandler) VerifyPublicShare(c echo.Context) error {
	var request verifyShareRequest
	if err := c.Bind(&request); err != nil {
		return BizError(model.ErrInvalidInput)
	}
	info, err := h.svc.VerifyPublicPassword(c.Request().Context(), c.Param("slug"), request.PasswordCode)
	if err != nil {
		return BizError(err)
	}
	return OK(c, info)
}

func (h *ShareHandler) ServePublicFile(c echo.Context) error {
	passwordCode := c.QueryParam("code")
	fileSlug := c.QueryParam("fileSlug")
	res, err := h.svc.OpenSharedFile(c.Request().Context(), c.Param("slug"), passwordCode, fileSlug)
	if err != nil {
		return BizError(err)
	}
	defer res.File.Close()

	stat, err := res.File.Stat()
	if err != nil {
		return fmt.Errorf("stat shared file: %w", err)
	}

	disposition := "inline"
	if c.QueryParam("download") == "1" || c.QueryParam("download") == "true" {
		disposition = "attachment"
	}
	resp := c.Response()
	resp.Header().Set(echo.HeaderContentDisposition, fmt.Sprintf(`%s; filename="%s"`, disposition, res.Name))
	resp.Header().Set(echo.HeaderContentType, res.MimeType)
	resp.Header().Set("Cache-Control", "public, max-age=3600")
	resp.Header().Set("ETag", `"`+res.FileHash+`"`)

	http.ServeContent(resp, c.Request(), res.Name, stat.ModTime(), res.File)
	return nil
}

func parseOptionalShareTime(value *string) (*time.Time, error) {
	if value == nil {
		return nil, nil
	}
	trimmed := strings.TrimSpace(*value)
	if trimmed == "" {
		return nil, nil
	}
	parsed, err := time.Parse(time.RFC3339, trimmed)
	if err != nil {
		return nil, BizError(model.ErrInvalidInput)
	}
	return &parsed, nil
}

func (r *updateShareRequest) UnmarshalJSON(data []byte) error {
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	if value, ok := raw["passwordCode"]; ok {
		r.PasswordCodeSet = true
		if string(value) != "null" {
			var passwordCode string
			if err := json.Unmarshal(value, &passwordCode); err != nil {
				return BizError(err)
			}
			r.PasswordCode = &passwordCode
		}
	}
	if value, ok := raw["expiresAt"]; ok {
		r.ExpiresAtSet = true
		if string(value) != "null" {
			var expiresAt string
			if err := json.Unmarshal(value, &expiresAt); err != nil {
				return BizError(err)
			}
			r.ExpiresAt = &expiresAt
		}
	}
	return nil
}
