package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/event-driven-analytics-engine/ingestion-api/kafka"
	"github.com/event-driven-analytics-engine/ingestion-api/telemetry"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
)

// EventPayload represents the incoming JSON payload.
type EventPayload struct {
	UserID    string `json:"user_id" binding:"required"`
	Action    string `json:"action" binding:"required"`
	Timestamp string `json:"timestamp" binding:"required"`
}

var dbpool *pgxpool.Pool

func main() {
	// Initialize OpenTelemetry Tracing
	tp, err := telemetry.InitTracer()
	if err != nil {
		log.Fatalf("Failed to initialize tracer: %v", err)
	}
	defer func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := tp.Shutdown(ctx); err != nil {
			log.Printf("Error shutting down tracer provider: %v", err)
		}
	}()

	// Set global W3C TraceContext propagator for distributed trace stitching.
	otel.SetTextMapPropagator(propagation.TraceContext{})

	// Initialize Kafka Producer
	if err := kafka.InitProducer(); err != nil {
		log.Fatalf("Failed to initialize Kafka producer: %v", err)
	}
	defer kafka.CloseProducer()

	// Initialize Database Connection
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		dbURL = "postgres://postgres:postgres@localhost:5433/analytics?sslmode=disable"
	}

	dbpool, err = pgxpool.New(context.Background(), dbURL)
	if err != nil {
		log.Fatalf("Unable to create connection pool: %v", err)
	}
	defer dbpool.Close()

	// Initialize Gin Router
	r := gin.Default()

	// Register OpenTelemetry tracing and latency measurement middlewares.
	// These MUST come before route definitions so every request is instrumented.
	r.Use(telemetry.TracingMiddleware())
	r.Use(telemetry.LatencyMiddleware())

	// Apply CORS middleware
	r.Use(func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	})

	// Health check
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "UP"})
	})

	v1 := r.Group("/api/v1")
	{
		v1.POST("/events", handleIngestEvent)
		v1.GET("/events/status/:user_id", handleCheckEventStatus)
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Ingestion API starting on port %s", port)
	if err := r.Run(":" + port); err != nil {
		log.Fatalf("Failed to run server: %v", err)
	}
}

func handleIngestEvent(c *gin.Context) {
	var event EventPayload

	// Validate JSON schema
	if err := c.ShouldBindJSON(&event); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid payload format. Expected {user_id, action, timestamp}"})
		return
	}

	// Push to message broker asynchronously
	if err := kafka.PublishEvent(event); err != nil {
		log.Printf("Failed to publish event to Kafka: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to queue event"})
		return
	}

	// Return 202 Accepted immediately setup
	c.JSON(http.StatusAccepted, gin.H{"status": "queued"})
}

func handleCheckEventStatus(c *gin.Context) {
	userID := c.Param("user_id")
	action := c.Query("action")

	if userID == "" || action == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing user_id or action parameter"})
		return
	}

	// Check if this event has arrived in the analytics_events table recently
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	var exists bool
	// We check for events from the last 5 minutes to avoid polling old data
	err := dbpool.QueryRow(ctx,
		"SELECT EXISTS(SELECT 1 FROM analytics_events WHERE user_id = $1 AND action = $2 AND processed_at > NOW() - INTERVAL '5 minutes')",
		userID, action).Scan(&exists)

	if err != nil {
		log.Printf("Database check error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not verify status"})
		return
	}

	if exists {
		c.JSON(http.StatusOK, gin.H{"status": "persisted"})
	} else {
		c.JSON(http.StatusOK, gin.H{"status": "queued"})
	}
}
