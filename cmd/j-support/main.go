package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	_ "github.com/AzizovHikmatullo/j-support/docs"
	"github.com/AzizovHikmatullo/j-support/internal/app"
	"github.com/AzizovHikmatullo/j-support/internal/config"
	"github.com/AzizovHikmatullo/j-support/internal/db"
	"github.com/AzizovHikmatullo/j-support/internal/ws"
	"github.com/AzizovHikmatullo/j-support/pkg/logger"
)

// @title           J-Support API
// @version         1.0
// @description     Система поддержки клиентов с тикетами, сценариями бота и мультиканальностью

// @contact.name   Hikmatullo Azizov

// @host      localhost:8080
// @BasePath  /

// @securityDefinitions.apikey Bearer
// @in header
// @name Authorization
// @description Type "Bearer" + JWT token
func main() {
	logger := logger.NewLogger()

	cfg, err := config.LoadConfig()
	if err != nil {
		logger.Error("failed to load config", slog.String("error", err.Error()))
		return
	}

	db, err := db.Connect(cfg)
	if err != nil {
		logger.Error("failed to connect to db", slog.String("error", err.Error()))
		return
	}

	hub := ws.NewHub()

	app := app.NewApp(cfg, logger, db, hub)

	app.Run()

	// GRACEFUL SHUTDOWN
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	<-quit

	logger.Info("Shutting down server...")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := app.Srv.Shutdown(shutdownCtx); err != nil {
		logger.Info("Server forced to shutdown", slog.String("error", err.Error()))
	}

	hub.Shutdown()

	db.Close()

	logger.Info("Server exiting. Goodbye!")
}
