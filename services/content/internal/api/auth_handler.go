package api

import (
	"net/http"

	"github.com/b25/services/content/internal/domain"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

var validate = validator.New()

// Register handles user registration
// @Summary Register a new user
// @Tags auth
// @Accept json
// @Produce json
// @Param input body domain.RegisterUserInput true "Registration input"
// @Success 201 {object} SuccessResponse{data=domain.User}
// @Failure 400 {object} ErrorResponse
// @Failure 409 {object} ErrorResponse
// @Router /api/v1/auth/register [post]
func (h *Handler) Register(c *gin.Context) {
	var input domain.RegisterUserInput
	if err := c.ShouldBindJSON(&input); err != nil {
		h.sendError(c, http.StatusBadRequest, err, "invalid request body")
		return
	}

	if err := validate.Struct(input); err != nil {
		h.sendError(c, http.StatusBadRequest, err, "validation failed")
		return
	}

	user, err := h.authService.Register(c.Request.Context(), input)
	if err != nil {
		if err == domain.ErrUserExists {
			h.sendError(c, http.StatusConflict, err, "user already exists")
			return
		}
		h.logger.Error("registration failed", err)
		h.sendError(c, http.StatusInternalServerError, err, "registration failed")
		return
	}

	h.sendSuccess(c, http.StatusCreated, user, "user registered successfully")
}

// Login handles user authentication
// @Summary Login
// @Tags auth
// @Accept json
// @Produce json
// @Param input body domain.LoginInput true "Login credentials"
// @Success 200 {object} SuccessResponse{data=domain.AuthResponse}
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Router /api/v1/auth/login [post]
func (h *Handler) Login(c *gin.Context) {
	var input domain.LoginInput
	if err := c.ShouldBindJSON(&input); err != nil {
		h.sendError(c, http.StatusBadRequest, err, "invalid request body")
		return
	}

	if err := validate.Struct(input); err != nil {
		h.sendError(c, http.StatusBadRequest, err, "validation failed")
		return
	}

	authResponse, err := h.authService.Login(c.Request.Context(), input)
	if err != nil {
		if err == domain.ErrInvalidCredentials {
			h.sendError(c, http.StatusUnauthorized, err, "invalid credentials")
			return
		}
		h.logger.Error("login failed", err)
		h.sendError(c, http.StatusInternalServerError, err, "login failed")
		return
	}

	h.sendSuccess(c, http.StatusOK, authResponse, "login successful")
}

// GetMe returns current user information
// @Summary Get current user
// @Tags auth
// @Security BearerAuth
// @Produce json
// @Success 200 {object} SuccessResponse{data=domain.User}
// @Failure 401 {object} ErrorResponse
// @Router /api/v1/auth/me [get]
func (h *Handler) GetMe(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		h.sendError(c, http.StatusUnauthorized, domain.ErrUnauthorized, "user not authenticated")
		return
	}

	user, err := h.authService.GetUserByID(c.Request.Context(), userID.(gin.H)["user_id"].(string))
	if err != nil {
		h.sendError(c, http.StatusInternalServerError, err, "failed to get user")
		return
	}

	h.sendSuccess(c, http.StatusOK, user, "")
}
