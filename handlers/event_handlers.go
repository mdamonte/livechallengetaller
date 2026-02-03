package handlers

import (
	"net/http"

	"livechallengetaller/models"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type EventHandler struct {
	repo *models.EventRepository
}

func NewEventHandler(repo *models.EventRepository) *EventHandler {
	return &EventHandler{repo: repo}
}

// CreateEvent handles POST /events
func (h *EventHandler) CreateEvent(c *gin.Context) {
	var event models.Event

	// Bind JSON payload to Event struct
	if err := c.ShouldBindJSON(&event); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Validate end time is after start time
	if event.EndTime.Before(event.StartTime) || event.EndTime.Equal(event.StartTime) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "end_time must be after start_time"})
		return
	}

	// Create the event
	if err := h.repo.Create(c.Request.Context(), &event); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create event"})
		return
	}

	c.JSON(http.StatusCreated, event)
}

// GetEvent handles GET /events/:id
func (h *EventHandler) GetEvent(c *gin.Context) {
	// Parse and validate UUID
	id := c.Param("id")
	if _, err := uuid.Parse(id); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid event ID format"})
		return
	}

	event, err := h.repo.GetByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch event"})
		return
	}

	if event == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "event not found"})
		return
	}

	c.JSON(http.StatusOK, event)
}

// ListEvents handles GET /events
func (h *EventHandler) ListEvents(c *gin.Context) {
	events, err := h.repo.List(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch events"})
		return
	}

	if events == nil {
		events = []*models.Event{} // Return empty array instead of null
	}

	c.JSON(http.StatusOK, events)
}
