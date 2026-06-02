package handler

import (
	"io"
	"strconv"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/labstack/echo/v4"
	"github.com/rs/zerolog"

	"github.com/netdisk/server/internal/model"
	"github.com/netdisk/server/internal/service"
)

type UploadHandler struct {
	svc    *service.UploadService
	logger zerolog.Logger
}

func NewUploadHandler(svc *service.UploadService, logger zerolog.Logger) *UploadHandler {
	return &UploadHandler{svc: svc, logger: logger}
}

func (h *UploadHandler) PreCheck(c echo.Context) error {
	userID, err := requireUserID(c)
	if err != nil {
		return err
	}

	var input service.PreCheckRequest
	if err := c.Bind(&input); err != nil {
		return model.ErrInvalidInput
	}

	resp, err := h.svc.PreCheck(c.Request().Context(), userID, input)
	if err != nil {
		return err
	}

	return OK(c, resp)
}

func (h *UploadHandler) RequestChallenge(c echo.Context) error {
	userID, err := requireUserID(c)
	if err != nil {
		return err
	}

	var input service.RequestChallengeRequest
	if err := c.Bind(&input); err != nil {
		return model.ErrInvalidInput
	}

	resp, err := h.svc.RequestChallenge(c.Request().Context(), userID, input)
	if err != nil {
		return err
	}

	return OK(c, resp)
}

func (h *UploadHandler) Verify(c echo.Context) error {
	userID, err := requireUserID(c)
	if err != nil {
		return err
	}

	var input service.VerifyRequest
	if err := c.Bind(&input); err != nil {
		return model.ErrInvalidInput
	}

	resp, err := h.svc.Verify(c.Request().Context(), userID, input)
	if err != nil {
		return err
	}

	return OK(c, resp)
}

func (h *UploadHandler) Init(c echo.Context) error {
	userID, err := requireUserID(c)
	if err != nil {
		return err
	}

	var input service.InitRequest
	if err := c.Bind(&input); err != nil {
		return model.ErrInvalidInput
	}

	resp, err := h.svc.Init(c.Request().Context(), userID, input)
	if err != nil {
		return err
	}

	return Created(c, resp)
}

func (h *UploadHandler) UploadChunk(c echo.Context) error {
	userID, err := requireUserID(c)
	if err != nil {
		return err
	}

	uploadSlug := c.FormValue("uploadSlug")
	chunkIndexStr := c.FormValue("chunkIndex")
	chunkIndex, err := strconv.Atoi(chunkIndexStr)
	if err != nil {
		h.logger.Warn().Int64("userID", userID).Str("uploadSlug", uploadSlug).Str("chunkIndex", chunkIndexStr).Int64("contentLength", c.Request().ContentLength).Msg("upload chunk: invalid chunk index")
		return model.ErrInvalidInput
	}

	file, err := c.FormFile("chunkData")
	if err != nil {
		h.logger.Warn().Int64("userID", userID).Str("uploadSlug", uploadSlug).Int("chunkIndex", chunkIndex).Int64("contentLength", c.Request().ContentLength).Err(err).Msg("upload chunk: missing or invalid chunk data")
		return model.ErrInvalidInput
	}

	f, err := file.Open()
	if err != nil {
		h.logger.Error().Int64("userID", userID).Str("uploadSlug", uploadSlug).Int("chunkIndex", chunkIndex).Int64("fileSize", file.Size).Err(err).Msg("upload chunk: open multipart file failed")
		return model.ErrInternal
	}
	defer f.Close()

	data, err := io.ReadAll(f)
	if err != nil {
		h.logger.Error().Int64("userID", userID).Str("uploadSlug", uploadSlug).Int("chunkIndex", chunkIndex).Int64("fileSize", file.Size).Err(err).Msg("upload chunk: read multipart file failed")
		return model.ErrInternal
	}

	h.logger.Debug().Int64("userID", userID).Str("uploadSlug", uploadSlug).Int("chunkIndex", chunkIndex).Int64("declaredSize", file.Size).Int("readSize", len(data)).Msg("upload chunk: received")

	if err := h.svc.AppendChunk(c.Request().Context(), userID, uploadSlug, int32(chunkIndex), data); err != nil {
		return err
	}

	return OK(c, map[string]string{"message": "chunk uploaded"})
}

func (h *UploadHandler) Complete(c echo.Context) error {
	userID, err := requireUserID(c)
	if err != nil {
		return err
	}

	var input struct {
		UploadSlug string `json:"uploadSlug"`
	}
	if err := c.Bind(&input); err != nil {
		return model.ErrInvalidInput
	}

	resp, err := h.svc.Complete(c.Request().Context(), userID, input.UploadSlug)
	if err != nil {
		return err
	}

	return OK(c, resp)
}

func (h *UploadHandler) UpdateHash(c echo.Context) error {
	userID, err := requireUserID(c)
	if err != nil {
		return err
	}

	var input service.UpdateHashRequest
	if err := c.Bind(&input); err != nil {
		return model.ErrInvalidInput
	}

	if err := h.svc.UpdateHash(c.Request().Context(), userID, input); err != nil {
		return err
	}

	return OK(c, map[string]string{"message": "hash updated"})
}

func (h *UploadHandler) GetStatus(c echo.Context) error {
	userID, err := requireUserID(c)
	if err != nil {
		return err
	}

	uploadSlug := c.Param("upload_slug")
	resp, err := h.svc.GetStatus(c.Request().Context(), userID, uploadSlug)
	if err != nil {
		return err
	}

	return OK(c, resp)
}

func (h *UploadHandler) ListTasks(c echo.Context) error {
	userID, err := requireUserID(c)
	if err != nil {
		return err
	}

	limit, _ := strconv.Atoi(c.QueryParam("limit"))
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	offset, _ := strconv.Atoi(c.QueryParam("offset"))
	if offset < 0 {
		offset = 0
	}

	var startDate, endDate pgtype.Timestamptz
	if v := c.QueryParam("start_date"); v != "" {
		if t, err := time.Parse(time.RFC3339, v); err == nil {
			startDate = pgtype.Timestamptz{Time: t, Valid: true}
		}
	}
	if v := c.QueryParam("end_date"); v != "" {
		if t, err := time.Parse(time.RFC3339, v); err == nil {
			endDate = pgtype.Timestamptz{Time: t, Valid: true}
		}
	}

	status := c.QueryParam("status")

	resp, err := h.svc.ListTasks(c.Request().Context(), userID, limit, offset, startDate, endDate, status)
	if err != nil {
		return err
	}

	return OK(c, resp)
}

func (h *UploadHandler) RetryTask(c echo.Context) error {
	userID, err := requireUserID(c)
	if err != nil {
		return err
	}

	slug := c.Param("slug")
	if slug == "" {
		return model.ErrInvalidInput
	}

	resp, err := h.svc.RetryTask(c.Request().Context(), userID, slug)
	if err != nil {
		return err
	}

	return Created(c, resp)
}

func (h *UploadHandler) DeleteTask(c echo.Context) error {
	userID, err := requireUserID(c)
	if err != nil {
		return err
	}

	slug := c.Param("slug")
	if slug == "" {
		return model.ErrInvalidInput
	}

	if err := h.svc.DeleteTask(c.Request().Context(), userID, slug); err != nil {
		return err
	}
	return c.NoContent(204)
}

func (h *UploadHandler) DeleteTasks(c echo.Context) error {
	userID, err := requireUserID(c)
	if err != nil {
		return err
	}

	var req struct {
		Slugs []string `json:"slugs"`
	}
	if err := c.Bind(&req); err != nil {
		return model.ErrInvalidInput
	}
	if len(req.Slugs) == 0 {
		return model.ErrInvalidInput
	}

	if err := h.svc.DeleteTasks(c.Request().Context(), userID, req.Slugs); err != nil {
		return err
	}
	return c.NoContent(204)
}
