package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/referendumApp/statusphere-example-app-go/internal/config"
	"github.com/referendumApp/statusphere-example-app-go/internal/db"
	"github.com/referendumApp/statusphere-example-app-go/internal/server"
	"github.com/joho/godotenv"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func main() {
	// Set up logging
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr, TimeFormat: time.RFC3339})

	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Info().Msg("No .env file found, using environment variables")
	}

	// Initialize configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to load configuration")
	}

	// Set log level based on environment
	logLevel := zerolog.InfoLevel
	if cfg.Debug {
		logLevel = zerolog.DebugLevel
	}
	zerolog.SetGlobalLevel(logLevel)

	// Initialize database
	database, err := db.New(cfg.DBPath)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to initialize database")
	}

	// Run migrations
	if err := database.Migrate(); err != nil {
		log.Fatal().Err(err).Msg("Failed to run database migrations")
	}

	// Create and initialize the server
	srv, err := server.New(cfg, database)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to create server")
	}

	// Start the server in a separate goroutine
	go func() {
		addr := fmt.Sprintf("%s:%d", cfg.Host, cfg.Port)
		log.Info().Str("addr", addr).Msg("Starting server")

		if err := srv.Start(); err != nil && err != http.ErrServerClosed {
			log.Fatal().Err(err).Msg("Server failed")
		}
	}()

	// Wait for interrupt signal to gracefully shut down the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Info().Msg("Shutting down server...")

	// Create a deadline for server shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal().Err(err).Msg("Server forced to shutdown")
	}

	log.Info().Msg("Server exited properly")
}