package main

import (
	"context"
	"database/sql"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"livechallengetaller/handlers"
	"livechallengetaller/middleware"
	"livechallengetaller/models"

	"github.com/gin-gonic/gin"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/joho/godotenv"
)

var db *sql.DB

func main() {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Printf("Warning: .env file not found, using environment variables")
	}

	// Initialize database connection
	initDB()
	defer db.Close()

	// Initialize database schema
	if err := models.InitSchema(db); err != nil {
		log.Fatalf("Failed to initialize database schema: %v", err)
	}

	// Create router with middleware
	r := gin.Default()
	r.Use(middleware.ErrorHandler())

	// Setup routes
	setupRoutes(r)

	// Start server
	srv := &http.Server{
		Addr:    ":8080",
		Handler: r,
	}

	// Graceful shutdown
	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server error: %v", err)
		}
	}()

	log.Println("Server started on :8080")

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal("Server forced to shutdown:", err)
	}

	log.Println("Server exiting")
}

func initDB() {
	var err error

	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		log.Fatal("DATABASE_URL environment variable is not set. Example: postgres://username:password@localhost:5432/events_db?sslmode=disable")
	}

	db, err = sql.Open("pgx", dbURL)
	if err != nil {
		log.Fatalf("Error connecting to database: %v", err)
	}

	// Test the database connection
	if err = db.Ping(); err != nil {
		log.Fatalf("Error pinging database: %v", err)
	}

	// Set connection pool settings
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(25)
	db.SetConnMaxLifetime(5 * time.Minute)

	log.Println("Successfully connected to database")
}

func setupRoutes(r *gin.Engine) {
	// Initialize repositories
	eventRepo := models.NewEventRepository(db)

	// Initialize handlers
	eh := handlers.NewEventHandler(eventRepo)

	// Health check
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	// API v1 routes
	v1 := r.Group("/api/v1")
	{
		v1.GET("/events", eh.ListEvents)
		v1.GET("/events/:id", eh.GetEvent)
		v1.POST("/events", eh.CreateEvent)
	}
}
