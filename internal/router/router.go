package router

import (
	"net/http"

	"github.com/irvanmhndra/jejak-api/internal/handler"
	"github.com/irvanmhndra/jejak-api/internal/service"
	"github.com/labstack/echo/v5"
)

func Register(
	e *echo.Echo,
	authSvc *service.AuthService,
	authH *handler.AuthHandler,
	activityH *handler.ActivityHandler,
	statsH *handler.StatsHandler,
	settingH *handler.SettingHandler,
) {
	e.GET("/health", func(c *echo.Context) error {
		return c.JSON(http.StatusOK, map[string]string{"status": "ok"})
	})

	v1 := e.Group("/api/v1")
	mid := handler.AuthMiddleware(authSvc)

	// Auth
	auth := v1.Group("/auth")
	auth.POST("/register", authH.Register)
	auth.POST("/login", authH.Login)
	auth.POST("/refresh", authH.Refresh)
	auth.POST("/logout", authH.Logout, mid)

	v1.GET("/me", authH.Me, mid)

	// Activities
	act := v1.Group("/activities", mid)
	act.POST("", activityH.Create)
	act.POST("/manual", activityH.CreateManual)
	act.POST("/import", activityH.Import)
	act.GET("", activityH.List)
	act.GET("/:id", activityH.Get)
	act.PUT("/:id", activityH.Update)
	act.DELETE("/:id", activityH.Delete)
	act.GET("/:id/gpx", activityH.ExportGPX)

	// Stats & records
	v1.GET("/stats/summary", statsH.Summary, mid)
	v1.GET("/records", statsH.Records, mid)

	// Settings
	settings := v1.Group("/settings", mid)
	settings.GET("", settingH.Get)
	settings.PUT("", settingH.Update)
}
