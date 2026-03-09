package main

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestHealthCheckRoute(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.Default()
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "UP"})
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/health", nil)
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("Expected status 200, got %d", w.Code)
	}
}

func TestEventsRouteInvalidPayload(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.Default()
	router.POST("/api/v1/events", handleIngestEvent)

	w := httptest.NewRecorder()
	// Invalid payload missing 'action'
	req, _ := http.NewRequest("POST", "/api/v1/events", bytes.NewBuffer([]byte(`{"user_id": "123", "timestamp": "2026-03-09T21:19:19Z"}`)))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	// Should fail schema validation before attempting Kafka publish
	if w.Code != http.StatusBadRequest {
		t.Fatalf("Expected status 400 for invalid payload, got %d", w.Code)
	}
}
