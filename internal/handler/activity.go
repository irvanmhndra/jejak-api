package handler

import (
	"io"
	"net/http"
	"strconv"

	"github.com/irvanmhndra/jejak-api/internal/repository"
	"github.com/irvanmhndra/jejak-api/internal/service"
	"github.com/irvanmhndra/jejak-api/pkg/httputil"
	"github.com/labstack/echo/v5"
)

type ActivityHandler struct {
	svc *service.ActivityService
}

func NewActivityHandler(svc *service.ActivityService) *ActivityHandler { return &ActivityHandler{svc} }

func (h *ActivityHandler) Create(c *echo.Context) error {
	var req service.CreateGPSRequest
	if ok, err := bindValidate(c, &req); !ok {
		return err
	}
	a, err := h.svc.CreateFromGPS(c.Request().Context(), getUserID(c), req)
	if err != nil {
		return respondErr(c, err)
	}
	return httputil.Success(c, http.StatusCreated, a)
}

func (h *ActivityHandler) CreateManual(c *echo.Context) error {
	var req service.CreateManualRequest
	if ok, err := bindValidate(c, &req); !ok {
		return err
	}
	a, err := h.svc.CreateManual(c.Request().Context(), getUserID(c), req)
	if err != nil {
		return respondErr(c, err)
	}
	return httputil.Success(c, http.StatusCreated, a)
}

func (h *ActivityHandler) Import(c *echo.Context) error {
	data, err := readGPXBody(c)
	if err != nil {
		return httputil.Error(c, http.StatusBadRequest, "gagal membaca file gpx", "BAD_REQUEST")
	}
	a, err := h.svc.ImportGPX(c.Request().Context(), getUserID(c), data, c.QueryParam("type"))
	if err != nil {
		return respondErr(c, err)
	}
	return httputil.Success(c, http.StatusCreated, a)
}

func (h *ActivityHandler) List(c *echo.Context) error {
	limit, _ := strconv.Atoi(c.QueryParam("limit"))
	offset, _ := strconv.Atoi(c.QueryParam("offset"))
	items, total, err := h.svc.List(c.Request().Context(), repository.ActivityListParams{
		UserID: getUserID(c),
		Type:   c.QueryParam("type"),
		Limit:  limit,
		Offset: offset,
	})
	if err != nil {
		return respondErr(c, err)
	}
	if limit <= 0 {
		limit = 20
	}
	return httputil.SuccessWithPagination(c, items, total, limit, offset)
}

func (h *ActivityHandler) Get(c *echo.Context) error {
	a, err := h.svc.Get(c.Request().Context(), c.Param("id"), getUserID(c))
	if err != nil {
		return respondErr(c, err)
	}
	return httputil.Success(c, http.StatusOK, a)
}

func (h *ActivityHandler) Update(c *echo.Context) error {
	var req service.UpdateActivityRequest
	if ok, err := bindValidate(c, &req); !ok {
		return err
	}
	a, err := h.svc.Update(c.Request().Context(), c.Param("id"), getUserID(c), req)
	if err != nil {
		return respondErr(c, err)
	}
	return httputil.Success(c, http.StatusOK, a)
}

func (h *ActivityHandler) Delete(c *echo.Context) error {
	if err := h.svc.Delete(c.Request().Context(), c.Param("id"), getUserID(c)); err != nil {
		return respondErr(c, err)
	}
	return httputil.SuccessMsg(c, http.StatusOK, "aktivitas dihapus")
}

func (h *ActivityHandler) ExportGPX(c *echo.Context) error {
	gpx, title, err := h.svc.ExportGPX(c.Request().Context(), c.Param("id"), getUserID(c))
	if err != nil {
		return respondErr(c, err)
	}
	c.Response().Header().Set("Content-Disposition", `attachment; filename="`+sanitizeFilename(title)+`.gpx"`)
	return c.Blob(http.StatusOK, "application/gpx+xml", []byte(gpx))
}

func readGPXBody(c *echo.Context) ([]byte, error) {
	// multipart form field "file", else raw body.
	if file, err := c.FormFile("file"); err == nil {
		src, err := file.Open()
		if err != nil {
			return nil, err
		}
		defer func() { _ = src.Close() }()
		return io.ReadAll(src)
	}
	defer func() { _ = c.Request().Body.Close() }()
	return io.ReadAll(c.Request().Body)
}

func sanitizeFilename(s string) string {
	out := make([]rune, 0, len(s))
	for _, r := range s {
		switch {
		case r >= 'a' && r <= 'z', r >= 'A' && r <= 'Z', r >= '0' && r <= '9':
			out = append(out, r)
		case r == ' ' || r == '-' || r == '_':
			out = append(out, '-')
		}
	}
	if len(out) == 0 {
		return "activity"
	}
	return string(out)
}
