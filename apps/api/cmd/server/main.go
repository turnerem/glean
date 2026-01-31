package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/glean/api/internal/handlers"
	"github.com/glean/api/internal/middleware"
	"github.com/glean/api/internal/repository"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func main() {
	// Setup logging
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})

	// Get config from env
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	databaseURL := os.Getenv("DATABASE_URL")

	// Initialize repository
	var repo repository.Repository
	var err error

	if databaseURL != "" {
		log.Info().Msg("Connecting to PostgreSQL...")
		repo, err = repository.NewPostgresRepository(databaseURL)
		if err != nil {
			log.Fatal().Err(err).Msg("Failed to connect to database")
		}
		log.Info().Msg("Connected to PostgreSQL")
	} else {
		log.Warn().Msg("DATABASE_URL not set, using in-memory storage (data will be lost on restart)")
		repo = repository.NewMemoryRepository()
	}
	defer repo.Close()

	// Set repository for handlers
	handlers.SetRepository(repo)

	// Setup Gin
	if os.Getenv("GIN_MODE") == "" {
		gin.SetMode(gin.ReleaseMode)
	}

	r := gin.New()
	r.Use(middleware.Logger())
	r.Use(middleware.Recovery())
	r.Use(middleware.CORS())

	// Health check
	r.GET("/health", func(c *gin.Context) {
		if err := repo.Ping(c.Request.Context()); err != nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{"status": "unhealthy", "error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	// API routes
	api := r.Group("/api/v1")
	{
		// Auth routes (no auth required)
		auth := api.Group("/auth")
		{
			auth.POST("/register", handlers.Register)
			auth.POST("/login", handlers.Login)
		}

		// Protected routes
		protected := api.Group("")
		protected.Use(middleware.Auth())
		{
			// Notes
			protected.GET("/notes", handlers.GetNotes)
			protected.POST("/notes", handlers.CreateNote)
			protected.PUT("/notes/:id", handlers.UpdateNote)
			protected.DELETE("/notes/:id", handlers.DeleteNote)

			// Tags
			protected.GET("/tags", handlers.GetTags)
			protected.POST("/tags", handlers.CreateTag)

			// Sync
			protected.POST("/sync", handlers.Sync)
		}
	}

	// Create server
	srv := &http.Server{
		Addr:    ":" + port,
		Handler: r,
	}

	// Start server in goroutine
	go func() {
		log.Info().Str("port", port).Msg("Starting server")
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal().Err(err).Msg("Failed to start server")
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Info().Msg("Shutting down server...")

	// Graceful shutdown with 5 second timeout
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal().Err(err).Msg("Server forced to shutdown")
	}

	log.Info().Msg("Server exited")
}
