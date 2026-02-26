package app

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/AzizovHikmatullo/j-support/internal/categories"
	"github.com/AzizovHikmatullo/j-support/internal/config"
	"github.com/AzizovHikmatullo/j-support/internal/middleware"
	"github.com/AzizovHikmatullo/j-support/internal/tickets"
	"github.com/AzizovHikmatullo/j-support/internal/ws"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
)

type App struct {
	Srv *http.Server

	cfg    *config.Config
	db     *sqlx.DB
	logger *slog.Logger
	router *gin.Engine
	hub    *ws.Hub
}

func NewApp(cfg *config.Config, logger *slog.Logger, db *sqlx.DB, hub *ws.Hub) *App {
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
		hub:    hub,
	}
}

func (a *App) Run() {
	a.InitRoutes()

	go func() {
		a.logger.Info("Running server", slog.String("port", a.cfg.Server.Port))
		if err := a.Srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			a.logger.Info("Failed to run server", slog.String("error", err.Error()))
		}
	}()
}

func (a *App) InitRoutes() {
	a.router.Use(cors.New(cors.Config{
		AllowOrigins: []string{"http://localhost:5500", "http://127.0.0.1:5500"},
		AllowMethods: []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders: []string{
			"Origin",
			"Content-Type",
			"Authorization",
		},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	a.router.Use(middleware.LoggerMiddleware(a.logger))

	publisher := ws.NewPublisher(a.hub)

	categoriesRepo := categories.NewRepository(a.db)
	categoriesService := categories.NewService(categoriesRepo)
	categoriesHandler := categories.NewHandler(categoriesService)

	categoriesRoutes := a.router.Group("/categories")
	categoriesRoutes.Use(middleware.AuthMiddleware(a.cfg.JWT.Secret))
	{
		categoriesRoutes.POST("", middleware.RequireRole("admin"), categoriesHandler.Create)
		categoriesRoutes.GET("", middleware.RequireRole("admin", "support", "user"), categoriesHandler.Get)
		categoriesRoutes.PUT("/:id", middleware.RequireRole("admin"), categoriesHandler.Update)
	}

	ticketsRepo := tickets.NewRepository(a.db)
	ticketsService := tickets.NewService(ticketsRepo, categoriesRepo, publisher)
	ticketsHandler := tickets.NewHandler(ticketsService)

	ticketsRoutes := a.router.Group("/tickets")
	ticketsRoutes.Use(middleware.AuthMiddleware(a.cfg.JWT.Secret))
	{
		ticketsRoutes.GET("", middleware.RequireRole("user", "support", "admin"), ticketsHandler.Get)
		ticketsRoutes.POST("", middleware.RequireRole("user"), ticketsHandler.Create)
		ticketsRoutes.GET(":id", middleware.RequireRole("user", "support", "admin"), ticketsHandler.GetByID)
		ticketsRoutes.PATCH(":id/assign", middleware.RequireRole("support", "admin"), ticketsHandler.ChangeAssigned)
		ticketsRoutes.PATCH(":id/status", middleware.RequireRole("user", "support", "admin"), ticketsHandler.ChangeStatus)

		ticketsRoutes.POST(":id/messages", middleware.RequireRole("user", "support", "admin"), ticketsHandler.CreateMessage)
		ticketsRoutes.GET(":id/messages", middleware.RequireRole("user", "support", "admin"), ticketsHandler.GetMessages)
	}

	wsHandler := ws.NewWSHandler(a.hub)

	wsRoutes := a.router.Group("/ws")
	{
		wsRoutes.GET("/tickets/:id", wsHandler.ServeTicketWS)
	}

	a.logger.Info("All routes created")
}
