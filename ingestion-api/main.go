package main

import (
	"log"
	"net/http"
	"os"

	"github.com/event-driven-analytics-engine/ingestion-api/kafka"
	"github.com/gin-gonic/gin"
)

// EventPayload represents the incoming JSON payload.
type EventPayload struct {
	UserID    string `json:"user_id" binding:"required"`
	Action    string `json:"action" binding:"required"`
	Timestamp string `json:"timestamp" binding:"required"`
}

func main() {
	// Initialize Kafka Producer
	if err := kafka.InitProducer(); err != nil {
		log.Fatalf("Failed to initialize Kafka producer: %v", err)
	}
	defer kafka.CloseProducer()

	// Initialize GinRouter
	r := gin.Default()

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
