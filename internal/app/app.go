package app

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/AzizovHikmatullo/j-support/internal/categories"
	"github.com/AzizovHikmatullo/j-support/internal/channel"
	channelApp "github.com/AzizovHikmatullo/j-support/internal/channel/app"
	channelTelegram "github.com/AzizovHikmatullo/j-support/internal/channel/telegram"
	channelWeb "github.com/AzizovHikmatullo/j-support/internal/channel/web"
	"github.com/AzizovHikmatullo/j-support/internal/config"
	"github.com/AzizovHikmatullo/j-support/internal/contacts"
	"github.com/AzizovHikmatullo/j-support/internal/middleware"
	"github.com/AzizovHikmatullo/j-support/internal/scenario"
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

		ReadTimeout:  60 * time.Second,
		WriteTimeout: 60 * time.Second,
		IdleTimeout:  60 * time.Second,
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
		AllowOrigins: a.cfg.CORS.AllowedOrigins,
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
		categoriesRoutes.PATCH("/:id", middleware.RequireRole("admin"), categoriesHandler.Update)
	}

	contactRepo := contacts.NewRepository(a.db)
	contactService := contacts.NewService(contactRepo)

	registry := channel.NewRegistry()
	registry.Register(channelApp.New(contactService))
	registry.Register(channelWeb.New(contactService))
	registry.Register(channelTelegram.New(contactService))

	ticketsRepo := tickets.NewRepository(a.db)
	ticketsService := tickets.NewService(ticketsRepo, categoriesRepo, publisher, nil)
	ticketsHandler := tickets.NewHandler(ticketsService, registry)

	scenarioRepository := scenario.NewRepository(a.db)
	scenarioService := scenario.NewService(scenarioRepository, ticketsService)

	ticketsService.SetScenarioService(scenarioService)

	clientRoutes := a.router.Group("/tickets")
	clientRoutes.Use(middleware.ChannelIdentityMiddleware(a.cfg.JWT.Secret))
	{
		clientRoutes.POST("", middleware.RequireRole("user"), ticketsHandler.Create)
		clientRoutes.GET("", middleware.RequireRole("user"), ticketsHandler.GetMine)
		clientRoutes.GET(":id", middleware.RequireRole("user"), ticketsHandler.GetMineByID)
		clientRoutes.POST(":id/messages", middleware.RequireRole("user"), ticketsHandler.CreateMessageByUser)
		clientRoutes.GET(":id/messages", middleware.RequireRole("user"), ticketsHandler.GetMessagesForUser)
	}

	supportRoutes := a.router.Group("/support/tickets")
	supportRoutes.Use(middleware.AuthMiddleware(a.cfg.JWT.Secret))
	{
		supportRoutes.GET("", middleware.RequireRole("support", "admin"), ticketsHandler.Get)
		supportRoutes.GET(":id", middleware.RequireRole("support", "admin"), ticketsHandler.GetByID)
		supportRoutes.PATCH(":id/assign", middleware.RequireRole("support", "admin"), ticketsHandler.ChangeAssigned)
		supportRoutes.PATCH(":id/status", middleware.RequireRole("support", "admin"), ticketsHandler.ChangeStatus)
		supportRoutes.POST(":id/messages", middleware.RequireRole("support", "admin"), ticketsHandler.CreateMessageBySupport)
		supportRoutes.GET(":id/messages", middleware.RequireRole("support", "admin"), ticketsHandler.GetMessagesForSupport)
	}

	wsHandler := ws.NewWSHandler(a.hub)
	wsRoutes := a.router.Group("/ws")
	{
		wsRoutes.GET("/tickets/:id", wsHandler.ServeTicketWS)
	}

	initHandler := channel.NewHandler(contactService)
	initRoutes := a.router.Group("/init")
	{
		initRoutes.POST("/web", initHandler.InitWeb)
		initRoutes.POST("/telegram", initHandler.InitTelegram)
	}

	a.logger.Info("All routes created")
}
