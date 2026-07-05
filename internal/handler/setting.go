package handler

import (
	"net/http"

	"github.com/irvanmhndra/jejak-api/internal/service"
	"github.com/irvanmhndra/jejak-api/pkg/httputil"
	"github.com/labstack/echo/v5"
)

type SettingHandler struct {
	svc *service.SettingService
}

func NewSettingHandler(svc *service.SettingService) *SettingHandler { return &SettingHandler{svc} }

func (h *SettingHandler) Get(c *echo.Context) error {
	res, err := h.svc.Get(c.Request().Context(), getUserID(c))
	if err != nil {
		return respondErr(c, err)
	}
	return httputil.Success(c, http.StatusOK, res)
}

func (h *SettingHandler) Update(c *echo.Context) error {
	var req service.UpdateSettingRequest
	if ok, err := bindValidate(c, &req); !ok {
		return err
	}
	res, err := h.svc.Update(c.Request().Context(), getUserID(c), req)
	if err != nil {
		return respondErr(c, err)
	}
	return httputil.Success(c, http.StatusOK, res)
}
