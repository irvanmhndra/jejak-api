package handler

import (
	"net/http"

	"github.com/irvanmhndra/jejak-api/internal/service"
	"github.com/irvanmhndra/jejak-api/pkg/httputil"
	"github.com/labstack/echo/v5"
)

type StatsHandler struct {
	svc *service.StatsService
}

func NewStatsHandler(svc *service.StatsService) *StatsHandler { return &StatsHandler{svc} }

func (h *StatsHandler) Summary(c *echo.Context) error {
	res, err := h.svc.Summary(c.Request().Context(), getUserID(c))
	if err != nil {
		return respondErr(c, err)
	}
	return httputil.Success(c, http.StatusOK, res)
}

func (h *StatsHandler) Records(c *echo.Context) error {
	res, err := h.svc.Records(c.Request().Context(), getUserID(c))
	if err != nil {
		return respondErr(c, err)
	}
	return httputil.Success(c, http.StatusOK, res)
}
