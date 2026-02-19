package app

import (
	"log/slog"
	"net/http"
	"os"

	"github.com/AzizovHikmatullo/j-support/internal/categories"
	"github.com/AzizovHikmatullo/j-support/internal/config"
	"github.com/AzizovHikmatullo/j-support/internal/middleware"
	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
)

type App struct {
	Srv *http.Server

	cfg    *config.Config
	db     *sqlx.DB
	logger *slog.Logger
	router *gin.Engine
}

func NewApp(cfg *config.Config, logger *slog.Logger, db *sqlx.DB) *App {
	router := gin.New()

	srv := &http.Server{
		Addr:    ":" + cfg.Server.Port,
		Handler: router,
	}

	return &App{
		Srv: srv,

		cfg:    cfg,
		db:     db,
		logger: logger,
		router: router,
	}
}

func (a *App) Run() {
	a.InitRoutes()

	go func() {
		a.logger.Info("Running server", slog.String("port", a.cfg.Server.Port))
		if err := a.Srv.ListenAndServe(); err != nil {
			a.logger.Info("Failed to run server", slog.String("error", err.Error()))
			os.Exit(1)
		}
	}()
}

func (a *App) InitRoutes() {
	a.router.Use(middleware.LoggerMiddleware(a.logger))

	categoriesRepo := categories.NewRepository(a.db)
	categoriesService := categories.NewService(categoriesRepo)
	categoriesHandler := categories.NewHandler(categoriesService)

	categoriesRoutes := a.router.Group("/categories")
	{
		categoriesRoutes.POST("", categoriesHandler.Create)
		categoriesRoutes.GET("", categoriesHandler.Get)
		categoriesRoutes.PUT(":id", categoriesHandler.Update)
	}

	a.logger.Info("All routes created")
}
