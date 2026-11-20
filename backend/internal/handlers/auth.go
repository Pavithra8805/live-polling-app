package handlers

import (
	"errors"
	"net/http"

	"live-polling-app/backend/internal/httpx"
	"live-polling-app/backend/internal/service"

	"github.com/gin-gonic/gin"
)

type AuthHandler struct {
	auth *service.AuthService
}

type authRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func NewAuthHandler(auth *service.AuthService) *AuthHandler {
	return &AuthHandler{auth: auth}
}

func (h *AuthHandler) Signup(c *gin.Context) {
	h.authenticate(c, true)
}

func (h *AuthHandler) Login(c *gin.Context) {
	h.authenticate(c, false)
}

func (h *AuthHandler) authenticate(c *gin.Context, signup bool) {
	var request authRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		httpx.Error(c, http.StatusBadRequest, "invalid_json", "request body must be valid JSON")
		return
	}

	var result service.AuthResult
	var err error
	if signup {
		result, err = h.auth.Signup(c.Request.Context(), request.Email, request.Password)
	} else {
		result, err = h.auth.Login(c.Request.Context(), request.Email, request.Password)
	}
	switch {
	case errors.Is(err, service.ErrInvalidInput):
		httpx.Error(c, http.StatusBadRequest, "invalid_input", "email and password do not meet the requirements")
	case errors.Is(err, service.ErrEmailAlreadyUsed):
		httpx.Error(c, http.StatusConflict, "email_taken", "an account with this email already exists")
	case errors.Is(err, service.ErrInvalidCredentials):
		httpx.Error(c, http.StatusUnauthorized, "invalid_credentials", "email or password is incorrect")
	case err != nil:
		httpx.Error(c, http.StatusInternalServerError, "internal_error", "an unexpected error occurred")
	default:
		status := http.StatusOK
		if signup {
			status = http.StatusCreated
		}
		c.JSON(status, gin.H{"token": result.Token, "user": result.User})
	}
}
