package app

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/irvanmhndra/jejak-api/config"
	"github.com/irvanmhndra/jejak-api/internal/handler"
	"github.com/irvanmhndra/jejak-api/internal/repository/postgres"
	"github.com/irvanmhndra/jejak-api/internal/router"
	"github.com/irvanmhndra/jejak-api/internal/service"
	"github.com/irvanmhndra/jejak-api/pkg/apperror"
	"github.com/irvanmhndra/jejak-api/pkg/httputil"
	customValidator "github.com/irvanmhndra/jejak-api/pkg/validator"
	"github.com/jmoiron/sqlx"
	"github.com/labstack/echo/v5"
	"github.com/labstack/echo/v5/middleware"
	_ "github.com/lib/pq"
)

type App struct {
	cfg  *config.Config
	db   *sqlx.DB
	echo *echo.Echo
}

func New(cfg *config.Config) (*App, error) {
	db, err := sqlx.Connect("postgres", cfg.Postgres.DSN())
	if err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(5 * time.Minute)

	m, err := migrate.New("file://migrations", cfg.Postgres.DSN())
	if err != nil {
		return nil, err
	}
	if err := m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return nil, err
	}
	slog.Info("Migrations applied")

	// Repositories
	userRepo := postgres.NewUserRepository(db)
	sessionRepo := postgres.NewUserSessionRepository(db)
	settingRepo := postgres.NewSettingRepository(db)
	activityRepo := postgres.NewActivityRepository(db)
	prRepo := postgres.NewPersonalRecordRepository(db)
	statsRepo := postgres.NewStatsRepository(db)
	goalRepo := postgres.NewGoalRepository(db)

	// Services
	authSvc := service.NewAuthService(userRepo, sessionRepo, settingRepo)
	activitySvc := service.NewActivityService(activityRepo, prRepo)
	statsSvc := service.NewStatsService(statsRepo, prRepo)
	settingSvc := service.NewSettingService(settingRepo)
	goalSvc := service.NewGoalService(goalRepo, statsRepo)

	// Handlers
	authH := handler.NewAuthHandler(authSvc)
	activityH := handler.NewActivityHandler(activitySvc)
	statsH := handler.NewStatsHandler(statsSvc)
	settingH := handler.NewSettingHandler(settingSvc)
	goalH := handler.NewGoalHandler(goalSvc)

	// Echo
	e := echo.New()
	e.Validator = customValidator.New()

	e.Use(middleware.RequestLoggerWithConfig(middleware.RequestLoggerConfig{
		LogLatency:  true,
		LogRemoteIP: true,
		LogMethod:   true,
		LogURI:      true,
		LogStatus:   true,
		HandleError: false,
		LogValuesFunc: func(c *echo.Context, v middleware.RequestLoggerValues) error {
			if v.Error == nil {
				slog.Info("request", "method", v.Method, "uri", v.URI, "status", v.Status, "latency", v.Latency)
			} else {
				slog.Error("request error", "method", v.Method, "uri", v.URI, "status", v.Status, "error", v.Error)
			}
			return nil
		},
	}))
	e.Use(middleware.Recover())
	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins: []string{"*"},
		AllowMethods: []string{http.MethodGet, http.MethodPost, http.MethodPut, http.MethodDelete, http.MethodOptions},
		AllowHeaders: []string{"Accept", "Authorization", "Content-Type", "X-Request-ID"},
	}))
	e.Use(middleware.RequestID())
	e.HTTPErrorHandler = customHTTPErrorHandler

	router.Register(e, authSvc, authH, activityH, statsH, settingH, goalH)

	return &App{cfg: cfg, db: db, echo: e}, nil
}

func (a *App) Run() {
	server := &http.Server{Addr: ":" + a.cfg.Server.Port, Handler: a.echo}

	go func() {
		slog.Info("Starting server", "port", a.cfg.Server.Port)
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			slog.Error("Server error", "error", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	slog.Info("Shutting down server...")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := server.Shutdown(ctx); err != nil {
		slog.Error("Server forced to shutdown", "error", err)
		os.Exit(1)
	}
	slog.Info("Server exited gracefully")
}

func (a *App) Close() {
	if a.db != nil {
		_ = a.db.Close()
	}
}

func customHTTPErrorHandler(c *echo.Context, err error) {
	if r, _ := echo.UnwrapResponse(c.Response()); r != nil && r.Committed {
		return
	}
	if ae, ok := err.(*apperror.AppError); ok {
		_ = httputil.Error(c, ae.Code, ae.Message, ae.ErrCode)
		return
	}
	var sc echo.HTTPStatusCoder
	if errors.As(err, &sc) {
		code := sc.StatusCode()
		_ = c.JSON(code, httputil.Response{Success: false, Message: http.StatusText(code), ErrorCode: http.StatusText(code)})
		return
	}
	slog.Error("unhandled error", "method", c.Request().Method, "path", c.Request().URL.Path, "error", err)
	_ = c.JSON(http.StatusInternalServerError, httputil.Response{Success: false, Message: "internal server error", ErrorCode: "INTERNAL_ERROR"})
}
