package handler

import (
	"net/http"
	"strings"

	"github.com/irvanmhndra/jejak-api/internal/service"
	"github.com/irvanmhndra/jejak-api/pkg/apperror"
	"github.com/irvanmhndra/jejak-api/pkg/httputil"
	"github.com/labstack/echo/v5"
)

const userIDKey = "userID"

func AuthMiddleware(authSvc *service.AuthService) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c *echo.Context) error {
			token := extractToken(c)
			if token == "" {
				return httputil.Error(c, http.StatusUnauthorized, "token diperlukan", "UNAUTHORIZED")
			}
			user, err := authSvc.ValidateToken(c.Request().Context(), token)
			if err != nil {
				if ae, ok := err.(*apperror.AppError); ok {
					return httputil.Error(c, ae.Code, ae.Message, ae.ErrCode)
				}
				return httputil.Error(c, http.StatusUnauthorized, "token tidak valid", "UNAUTHORIZED")
			}
			if user == nil {
				return httputil.Error(c, http.StatusUnauthorized, "token tidak valid", "UNAUTHORIZED")
			}
			c.Set(userIDKey, user.ID)
			return next(c)
		}
	}
}

func getUserID(c *echo.Context) string {
	id, _ := c.Get(userIDKey).(string)
	return id
}

func extractToken(c *echo.Context) string {
	auth := c.Request().Header.Get("Authorization")
	if strings.HasPrefix(auth, "Bearer ") {
		return strings.TrimPrefix(auth, "Bearer ")
	}
	return ""
}
