package api

import (
	"net/http"
	"strconv"

	"github.com/b25/services/configuration/internal/domain"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

// CreateConfiguration handles configuration creation
func (h *Handler) CreateConfiguration(c *gin.Context) {
	var req domain.CreateConfigurationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Success: false,
			Error:   err.Error(),
			Code:    "INVALID_REQUEST",
		})
		return
	}

	// Get actor info from context (set by auth middleware)
	actorName := c.GetString("actor_name")
	if actorName == "" {
		actorName = req.CreatedBy
	}

	config, err := h.configService.Create(
		c.Request.Context(),
		&req,
		actorName,
		c.ClientIP(),
		c.Request.UserAgent(),
	)

	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusCreated, Response{
		Success: true,
		Data:    config,
		Message: "Configuration created successfully",
	})
}

// GetConfiguration handles retrieving a configuration by ID
func (h *Handler) GetConfiguration(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Success: false,
			Error:   "Invalid configuration ID",
			Code:    "INVALID_ID",
		})
		return
	}

	config, err := h.configService.GetByID(c.Request.Context(), id)
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, Response{
		Success: true,
		Data:    config,
	})
}

// GetConfigurationByKey handles retrieving a configuration by key
func (h *Handler) GetConfigurationByKey(c *gin.Context) {
	key := c.Param("key")
	if key == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Success: false,
			Error:   "Configuration key is required",
			Code:    "MISSING_KEY",
		})
		return
	}

	config, err := h.configService.GetByKey(c.Request.Context(), key)
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, Response{
		Success: true,
		Data:    config,
	})
}

// ListConfigurations handles listing configurations with filters
func (h *Handler) ListConfigurations(c *gin.Context) {
	filter := domain.ConfigurationFilter{}

	// Parse type filter
	if typeStr := c.Query("type"); typeStr != "" {
		configType := domain.ConfigType(typeStr)
		filter.Type = &configType
	}

	// Parse active filter
	if activeStr := c.Query("active"); activeStr != "" {
		isActive := activeStr == "true"
		filter.IsActive = &isActive
	}

	// Parse limit and offset
	if limitStr := c.Query("limit"); limitStr != "" {
		if limit, err := strconv.Atoi(limitStr); err == nil {
			filter.Limit = limit
		}
	} else {
		filter.Limit = 50
	}

	if offsetStr := c.Query("offset"); offsetStr != "" {
		if offset, err := strconv.Atoi(offsetStr); err == nil {
			filter.Offset = offset
		}
	}

	configs, err := h.configService.List(c.Request.Context(), filter)
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, PaginatedResponse{
		Success: true,
		Data:    configs,
		Total:   len(configs),
		Limit:   filter.Limit,
		Offset:  filter.Offset,
	})
}

// UpdateConfiguration handles updating a configuration
func (h *Handler) UpdateConfiguration(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Success: false,
			Error:   "Invalid configuration ID",
			Code:    "INVALID_ID",
		})
		return
	}

	var req domain.UpdateConfigurationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Success: false,
			Error:   err.Error(),
			Code:    "INVALID_REQUEST",
		})
		return
	}

	actorName := c.GetString("actor_name")
	if actorName == "" {
		actorName = req.UpdatedBy
	}

	config, err := h.configService.Update(
		c.Request.Context(),
		id,
		&req,
		actorName,
		c.ClientIP(),
		c.Request.UserAgent(),
	)

	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, Response{
		Success: true,
		Data:    config,
		Message: "Configuration updated successfully",
	})
}

// ActivateConfiguration handles activating a configuration
func (h *Handler) ActivateConfiguration(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Success: false,
			Error:   "Invalid configuration ID",
			Code:    "INVALID_ID",
		})
		return
	}

	actorID := c.GetString("actor_id")
	actorName := c.GetString("actor_name")
	if actorID == "" {
		actorID = "system"
	}
	if actorName == "" {
		actorName = "System"
	}

	err = h.configService.Activate(
		c.Request.Context(),
		id,
		actorID,
		actorName,
		c.ClientIP(),
		c.Request.UserAgent(),
	)

	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, Response{
		Success: true,
		Message: "Configuration activated successfully",
	})
}

// DeactivateConfiguration handles deactivating a configuration
func (h *Handler) DeactivateConfiguration(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Success: false,
			Error:   "Invalid configuration ID",
			Code:    "INVALID_ID",
		})
		return
	}

	actorID := c.GetString("actor_id")
	actorName := c.GetString("actor_name")
	if actorID == "" {
		actorID = "system"
	}
	if actorName == "" {
		actorName = "System"
	}

	err = h.configService.Deactivate(
		c.Request.Context(),
		id,
		actorID,
		actorName,
		c.ClientIP(),
		c.Request.UserAgent(),
	)

	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, Response{
		Success: true,
		Message: "Configuration deactivated successfully",
	})
}

// DeleteConfiguration handles deleting a configuration
func (h *Handler) DeleteConfiguration(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Success: false,
			Error:   "Invalid configuration ID",
			Code:    "INVALID_ID",
		})
		return
	}

	actorID := c.GetString("actor_id")
	actorName := c.GetString("actor_name")
	if actorID == "" {
		actorID = "system"
	}
	if actorName == "" {
		actorName = "System"
	}

	err = h.configService.Delete(
		c.Request.Context(),
		id,
		actorID,
		actorName,
		c.ClientIP(),
		c.Request.UserAgent(),
	)

	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, Response{
		Success: true,
		Message: "Configuration deleted successfully",
	})
}

// GetVersions handles retrieving all versions of a configuration
func (h *Handler) GetVersions(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Success: false,
			Error:   "Invalid configuration ID",
			Code:    "INVALID_ID",
		})
		return
	}

	versions, err := h.configService.GetVersions(c.Request.Context(), id)
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, Response{
		Success: true,
		Data:    versions,
	})
}

// RollbackConfiguration handles rolling back a configuration to a specific version
func (h *Handler) RollbackConfiguration(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Success: false,
			Error:   "Invalid configuration ID",
			Code:    "INVALID_ID",
		})
		return
	}

	var req domain.RollbackRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Success: false,
			Error:   err.Error(),
			Code:    "INVALID_REQUEST",
		})
		return
	}

	actorName := c.GetString("actor_name")
	if actorName == "" {
		actorName = req.RolledBackBy
	}

	config, err := h.configService.Rollback(
		c.Request.Context(),
		id,
		&req,
		actorName,
		c.ClientIP(),
		c.Request.UserAgent(),
	)

	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, Response{
		Success: true,
		Data:    config,
		Message: "Configuration rolled back successfully",
	})
}

// GetAuditLogs handles retrieving audit logs for a configuration
func (h *Handler) GetAuditLogs(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Success: false,
			Error:   "Invalid configuration ID",
			Code:    "INVALID_ID",
		})
		return
	}

	limit := 100
	if limitStr := c.Query("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil {
			limit = l
		}
	}

	logs, err := h.configService.GetAuditLogs(c.Request.Context(), id, limit)
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, Response{
		Success: true,
		Data:    logs,
	})
}

// handleError handles errors and returns appropriate HTTP responses
func (h *Handler) handleError(c *gin.Context, err error) {
	h.logger.Error("API error", zap.Error(err))

	switch err {
	case domain.ErrNotFound:
		c.JSON(http.StatusNotFound, ErrorResponse{
			Success: false,
			Error:   "Configuration not found",
			Code:    "NOT_FOUND",
		})
	case domain.ErrDuplicateKey:
		c.JSON(http.StatusConflict, ErrorResponse{
			Success: false,
			Error:   "Configuration key already exists",
			Code:    "DUPLICATE_KEY",
		})
	case domain.ErrInvalidVersion:
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Success: false,
			Error:   "Invalid configuration version",
			Code:    "INVALID_VERSION",
		})
	case domain.ErrValidationFailed:
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Success: false,
			Error:   err.Error(),
			Code:    "VALIDATION_FAILED",
		})
	case domain.ErrInvalidInput:
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Success: false,
			Error:   err.Error(),
			Code:    "INVALID_INPUT",
		})
	default:
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Success: false,
			Error:   "Internal server error",
			Code:    "INTERNAL_ERROR",
		})
	}
}
