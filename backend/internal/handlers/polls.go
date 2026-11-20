package handlers

import (
	"errors"
	"net/http"
	"strings"
	"time"

	"live-polling-app/backend/internal/httpx"
	"live-polling-app/backend/internal/repository"
	"live-polling-app/backend/internal/service"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/v2/bson"
)

type PollHandler struct {
	polls *service.PollService
}

type createPollRequest struct {
	Slug     string     `json:"slug"`
	Question string     `json:"question"`
	Options  []string   `json:"options"`
	ClosesAt *time.Time `json:"closesAt"`
}

func NewPollHandler(polls *service.PollService) *PollHandler {
	return &PollHandler{polls: polls}
}

func (h *PollHandler) Create(c *gin.Context) {
	var request createPollRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		httpx.Error(c, http.StatusBadRequest, "invalid_json", "request body must be valid JSON")
		return
	}
	ownerID, ok := userObjectID(c)
	if !ok {
		return
	}
	poll, err := h.polls.Create(c.Request.Context(), ownerID, service.CreatePollInput{
		Slug: request.Slug, Question: request.Question, Options: request.Options, ClosesAt: request.ClosesAt,
	})
	if err != nil {
		h.writePollError(c, err)
		return
	}
	c.JSON(http.StatusCreated, poll)
}

func (h *PollHandler) ListMine(c *gin.Context) {
	ownerID, ok := userObjectID(c)
	if !ok {
		return
	}
	polls, err := h.polls.ListMine(c.Request.Context(), ownerID)
	if err != nil {
		httpx.Error(c, http.StatusInternalServerError, "internal_error", "an unexpected error occurred")
		return
	}
	c.JSON(http.StatusOK, gin.H{"polls": polls})
}

func (h *PollHandler) GetPublic(c *gin.Context) {
	poll, err := h.polls.PublicBySlug(c.Request.Context(), c.Param("slug"))
	if errors.Is(err, repository.ErrPollNotFound) {
		httpx.Error(c, http.StatusNotFound, "poll_not_found", "poll was not found")
		return
	}
	if err != nil {
		httpx.Error(c, http.StatusInternalServerError, "internal_error", "an unexpected error occurred")
		return
	}
	c.JSON(http.StatusOK, poll)
}

func (h *PollHandler) Close(c *gin.Context) {
	h.manage(c, false)
}

func (h *PollHandler) Delete(c *gin.Context) {
	h.manage(c, true)
}

func (h *PollHandler) manage(c *gin.Context, deletePoll bool) {
	ownerID, ok := userObjectID(c)
	if !ok {
		return
	}
	var err error
	if deletePoll {
		err = h.polls.Delete(c.Request.Context(), ownerID, c.Param("slug"))
	} else {
		err = h.polls.Close(c.Request.Context(), ownerID, c.Param("slug"))
	}
	if errors.Is(err, repository.ErrPollNotFound) {
		httpx.Error(c, http.StatusNotFound, "poll_not_found", "poll was not found or is not owned by you")
		return
	}
	if err != nil {
		httpx.Error(c, http.StatusInternalServerError, "internal_error", "an unexpected error occurred")
		return
	}
	if deletePoll {
		c.Status(http.StatusNoContent)
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "closed"})
}

func (h *PollHandler) writePollError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, service.ErrPollInvalidInput):
		httpx.Error(c, http.StatusBadRequest, "invalid_poll", "poll fields do not meet the requirements")
	case errors.Is(err, repository.ErrSlugTaken):
		httpx.Error(c, http.StatusConflict, "slug_taken", "that poll slug is already in use")
	default:
		httpx.Error(c, http.StatusInternalServerError, "internal_error", "an unexpected error occurred")
	}
}

func userObjectID(c *gin.Context) (bson.ObjectID, bool) {
	value, exists := c.Get("userID")
	if !exists {
		httpx.Error(c, http.StatusUnauthorized, "missing_user", "authenticated user context is missing")
		return bson.NilObjectID, false
	}
	userIDString, ok := value.(string)
	if !ok {
		httpx.Error(c, http.StatusUnauthorized, "invalid_user", "authenticated user context is invalid")
		return bson.NilObjectID, false
	}
	userID, err := bson.ObjectIDFromHex(strings.TrimSpace(userIDString))
	if err != nil {
		httpx.Error(c, http.StatusUnauthorized, "invalid_user", "authenticated user context is invalid")
		return bson.NilObjectID, false
	}
	return userID, true
}
