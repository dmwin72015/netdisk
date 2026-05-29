package handler

import (
	"io"
	"strconv"

	"github.com/labstack/echo/v4"

	"github.com/netdisk/server/internal/service"
)

type UploadHandler struct {
	svc *service.UploadService
}

func NewUploadHandler(svc *service.UploadService) *UploadHandler {
	return &UploadHandler{svc: svc}
}

func (h *UploadHandler) PreCheck(c echo.Context) error {
	userID, err := requireUserID(c)
	if err != nil {
		return err
	}

	var input service.PreCheckRequest
	if err := c.Bind(&input); err != nil {
		return echo.NewHTTPError(400, "invalid request body")
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
		return echo.NewHTTPError(400, "invalid request body")
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
		return echo.NewHTTPError(400, "invalid request body")
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
		return echo.NewHTTPError(400, "invalid request body")
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
		return echo.NewHTTPError(400, "invalid chunk_index")
	}

	file, err := c.FormFile("chunkData")
	if err != nil {
		return echo.NewHTTPError(400, "chunk_data is required")
	}

	f, err := file.Open()
	if err != nil {
		return echo.NewHTTPError(400, "cannot open file")
	}
	defer f.Close()

	data, err := io.ReadAll(f)
	if err != nil {
		return echo.NewHTTPError(400, "cannot read file")
	}

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
		return echo.NewHTTPError(400, "invalid request body")
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
		return echo.NewHTTPError(400, "invalid request body")
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
