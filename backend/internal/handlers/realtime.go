package handlers

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"live-polling-app/backend/internal/httpx"
	"live-polling-app/backend/internal/realtime"
	"live-polling-app/backend/internal/repository"
	"live-polling-app/backend/internal/service"

	"github.com/gin-gonic/gin"
)

const voterCookie = "voter_id"

type RealtimeHandler struct {
	votes        *service.VoteService
	hub          *realtime.Hub
	cookieSecure bool
}

type voteRequest struct {
	OptionID string `json:"optionId"`
}

func NewRealtimeHandler(votes *service.VoteService, hub *realtime.Hub, cookieSecure bool) *RealtimeHandler {
	return &RealtimeHandler{votes: votes, hub: hub, cookieSecure: cookieSecure}
}

func (h *RealtimeHandler) GetResults(c *gin.Context) {
	results, err := h.votes.Results(c.Request.Context(), c.Param("slug"))
	if err != nil {
		h.writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, results)
}

func (h *RealtimeHandler) Vote(c *gin.Context) {
	var request voteRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		httpx.Error(c, http.StatusBadRequest, "invalid_json", "request body must be valid JSON")
		return
	}
	voterID, err := h.voterID(c)
	if err != nil {
		httpx.Error(c, http.StatusInternalServerError, "voter_id_error", "could not create a voter identity")
		return
	}
	results, err := h.votes.Vote(c.Request.Context(), c.Param("slug"), request.OptionID, voterID)
	if err != nil {
		h.writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, results)
}

func (h *RealtimeHandler) Stream(c *gin.Context) {
	results, err := h.votes.Results(c.Request.Context(), c.Param("slug"))
	if err != nil {
		h.writeError(c, err)
		return
	}
	pollID := results.Poll.ID.Hex()
	client := h.hub.Register(pollID)
	defer h.hub.Unregister(pollID, client)

	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")
	c.Header("X-Accel-Buffering", "no")

	writer := c.Writer
	initialPayload, err := marshalJSON(results)
	if err != nil {
		httpx.Error(c, http.StatusInternalServerError, "serialization_error", "could not serialize poll results")
		return
	}
	if _, err := fmt.Fprintf(writer, "event: results\ndata: %s\n\n", initialPayload); err != nil {
		return
	}
	writer.Flush()

	heartbeat := time.NewTicker(15 * time.Second)
	defer heartbeat.Stop()
	for {
		select {
		case <-c.Request.Context().Done():
			return
		case payload, ok := <-client:
			if !ok {
				return
			}
			if _, err := fmt.Fprintf(writer, "event: results\ndata: %s\n\n", payload); err != nil {
				return
			}
			writer.Flush()
		case <-heartbeat.C:
			if _, err := fmt.Fprint(writer, ": heartbeat\n\n"); err != nil {
				return
			}
			writer.Flush()
		}
	}
}

func (h *RealtimeHandler) voterID(c *gin.Context) (string, error) {
	if value, err := c.Cookie(voterCookie); err == nil && value != "" {
		return value, nil
	}
	buffer := make([]byte, 16)
	if _, err := rand.Read(buffer); err != nil {
		return "", err
	}
	value := hex.EncodeToString(buffer)
	sameSite := http.SameSiteLaxMode
	if h.cookieSecure {
		sameSite = http.SameSiteNoneMode
	}
	http.SetCookie(c.Writer, &http.Cookie{
		Name:     voterCookie,
		Value:    value,
		Path:     "/",
		MaxAge:   365 * 24 * 60 * 60,
		HttpOnly: true,
		Secure:   h.cookieSecure,
		SameSite: sameSite,
	})
	return value, nil
}

func (h *RealtimeHandler) writeError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, repository.ErrPollNotFound):
		httpx.Error(c, http.StatusNotFound, "poll_not_found", "poll was not found")
	case errors.Is(err, service.ErrPollClosed):
		httpx.Error(c, http.StatusConflict, "poll_closed", "poll is closed or expired")
	case errors.Is(err, service.ErrVoteInvalid):
		httpx.Error(c, http.StatusBadRequest, "invalid_vote", "selected option is not valid")
	case errors.Is(err, realtime.ErrDuplicateVoter):
		httpx.Error(c, http.StatusConflict, "already_voted", "this voter has already voted")
	default:
		httpx.Error(c, http.StatusInternalServerError, "internal_error", "an unexpected error occurred")
	}
}

func marshalJSON(value interface{}) (string, error) {
	data, err := json.Marshal(value)
	if err != nil {
		return "", err
	}
	return string(data), nil
}
