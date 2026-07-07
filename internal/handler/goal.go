package handler

import (
	"net/http"

	"github.com/irvanmhndra/jejak-api/internal/service"
	"github.com/irvanmhndra/jejak-api/pkg/httputil"
	"github.com/labstack/echo/v5"
)

type GoalHandler struct {
	svc *service.GoalService
}

func NewGoalHandler(svc *service.GoalService) *GoalHandler { return &GoalHandler{svc} }

func (h *GoalHandler) Create(c *echo.Context) error {
	var req service.CreateGoalRequest
	if ok, err := bindValidate(c, &req); !ok {
		return err
	}
	g, err := h.svc.Create(c.Request().Context(), getUserID(c), req)
	if err != nil {
		return respondErr(c, err)
	}
	return httputil.Success(c, http.StatusCreated, g)
}

func (h *GoalHandler) List(c *echo.Context) error {
	goals, err := h.svc.List(c.Request().Context(), getUserID(c))
	if err != nil {
		return respondErr(c, err)
	}
	return httputil.Success(c, http.StatusOK, goals)
}

func (h *GoalHandler) Delete(c *echo.Context) error {
	if err := h.svc.Delete(c.Request().Context(), c.Param("id"), getUserID(c)); err != nil {
		return respondErr(c, err)
	}
	return httputil.SuccessMsg(c, http.StatusOK, "target dihapus")
}
