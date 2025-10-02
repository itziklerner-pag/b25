package api

import (
	"net/http"

	"github.com/b25/services/content/internal/domain"
	"github.com/b25/services/content/internal/middleware"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// CreateContent creates a new content item
// @Summary Create content
// @Tags content
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param input body domain.CreateContentInput true "Content input"
// @Success 201 {object} SuccessResponse{data=domain.Content}
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Router /api/v1/content [post]
func (h *Handler) CreateContent(c *gin.Context) {
	userID := middleware.MustGetUserID(c)
	if userID == uuid.Nil {
		return
	}

	var input domain.CreateContentInput
	if err := c.ShouldBindJSON(&input); err != nil {
		h.sendError(c, http.StatusBadRequest, err, "invalid request body")
		return
	}

	if err := validate.Struct(input); err != nil {
		h.sendError(c, http.StatusBadRequest, err, "validation failed")
		return
	}

	content, err := h.contentService.CreateContent(c.Request.Context(), input, userID)
	if err != nil {
		if err == domain.ErrContentExists {
			h.sendError(c, http.StatusConflict, err, "content with this slug already exists")
			return
		}
		if err == domain.ErrForbidden {
			h.sendError(c, http.StatusForbidden, err, "insufficient permissions")
			return
		}
		h.logger.Error("failed to create content", err)
		h.sendError(c, http.StatusInternalServerError, err, "failed to create content")
		return
	}

	h.sendSuccess(c, http.StatusCreated, content, "content created successfully")
}

// GetContent retrieves content by ID
// @Summary Get content by ID
// @Tags content
// @Produce json
// @Param id path string true "Content ID"
// @Success 200 {object} SuccessResponse{data=domain.Content}
// @Failure 404 {object} ErrorResponse
// @Router /api/v1/content/{id} [get]
func (h *Handler) GetContent(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		h.sendError(c, http.StatusBadRequest, err, "invalid content ID")
		return
	}

	content, err := h.contentService.GetContent(c.Request.Context(), id)
	if err != nil {
		if err == domain.ErrContentNotFound {
			h.sendError(c, http.StatusNotFound, err, "content not found")
			return
		}
		h.sendError(c, http.StatusInternalServerError, err, "failed to get content")
		return
	}

	// Increment view count asynchronously
	go h.contentService.IncrementViewCount(c.Request.Context(), id)

	h.sendSuccess(c, http.StatusOK, content, "")
}

// GetContentBySlug retrieves content by slug
// @Summary Get content by slug
// @Tags content
// @Produce json
// @Param slug path string true "Content slug"
// @Success 200 {object} SuccessResponse{data=domain.Content}
// @Failure 404 {object} ErrorResponse
// @Router /api/v1/content/slug/{slug} [get]
func (h *Handler) GetContentBySlug(c *gin.Context) {
	slug := c.Param("slug")

	content, err := h.contentService.GetContentBySlug(c.Request.Context(), slug)
	if err != nil {
		if err == domain.ErrContentNotFound {
			h.sendError(c, http.StatusNotFound, err, "content not found")
			return
		}
		h.sendError(c, http.StatusInternalServerError, err, "failed to get content")
		return
	}

	// Increment view count asynchronously
	go h.contentService.IncrementViewCount(c.Request.Context(), content.ID)

	h.sendSuccess(c, http.StatusOK, content, "")
}

// UpdateContent updates content
// @Summary Update content
// @Tags content
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path string true "Content ID"
// @Param input body domain.UpdateContentInput true "Update input"
// @Success 200 {object} SuccessResponse{data=domain.Content}
// @Failure 400 {object} ErrorResponse
// @Failure 403 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Router /api/v1/content/{id} [put]
func (h *Handler) UpdateContent(c *gin.Context) {
	userID := middleware.MustGetUserID(c)
	if userID == uuid.Nil {
		return
	}

	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		h.sendError(c, http.StatusBadRequest, err, "invalid content ID")
		return
	}

	var input domain.UpdateContentInput
	if err := c.ShouldBindJSON(&input); err != nil {
		h.sendError(c, http.StatusBadRequest, err, "invalid request body")
		return
	}

	if err := validate.Struct(input); err != nil {
		h.sendError(c, http.StatusBadRequest, err, "validation failed")
		return
	}

	content, err := h.contentService.UpdateContent(c.Request.Context(), id, input, userID)
	if err != nil {
		if err == domain.ErrContentNotFound {
			h.sendError(c, http.StatusNotFound, err, "content not found")
			return
		}
		if err == domain.ErrForbidden {
			h.sendError(c, http.StatusForbidden, err, "insufficient permissions")
			return
		}
		h.sendError(c, http.StatusInternalServerError, err, "failed to update content")
		return
	}

	h.sendSuccess(c, http.StatusOK, content, "content updated successfully")
}

// DeleteContent deletes content
// @Summary Delete content
// @Tags content
// @Security BearerAuth
// @Param id path string true "Content ID"
// @Success 200 {object} SuccessResponse
// @Failure 403 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Router /api/v1/content/{id} [delete]
func (h *Handler) DeleteContent(c *gin.Context) {
	userID := middleware.MustGetUserID(c)
	if userID == uuid.Nil {
		return
	}

	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		h.sendError(c, http.StatusBadRequest, err, "invalid content ID")
		return
	}

	err = h.contentService.DeleteContent(c.Request.Context(), id, userID)
	if err != nil {
		if err == domain.ErrContentNotFound {
			h.sendError(c, http.StatusNotFound, err, "content not found")
			return
		}
		if err == domain.ErrForbidden {
			h.sendError(c, http.StatusForbidden, err, "insufficient permissions")
			return
		}
		h.sendError(c, http.StatusInternalServerError, err, "failed to delete content")
		return
	}

	h.sendSuccess(c, http.StatusOK, nil, "content deleted successfully")
}

// SearchContent searches and filters content
// @Summary Search content
// @Tags content
// @Produce json
// @Param query query string false "Search query"
// @Param type query string false "Content type"
// @Param status query string false "Content status"
// @Param author_id query string false "Author ID"
// @Param tags query []string false "Tags"
// @Param categories query []string false "Categories"
// @Param sort_by query string false "Sort by field"
// @Param sort_order query string false "Sort order (asc/desc)"
// @Param page query int false "Page number"
// @Param page_size query int false "Page size"
// @Success 200 {object} SuccessResponse{data=domain.PaginatedContent}
// @Router /api/v1/content [get]
func (h *Handler) SearchContent(c *gin.Context) {
	var params domain.SearchContentParams
	if err := c.ShouldBindQuery(&params); err != nil {
		h.sendError(c, http.StatusBadRequest, err, "invalid query parameters")
		return
	}

	result, err := h.contentService.SearchContent(c.Request.Context(), params)
	if err != nil {
		h.sendError(c, http.StatusInternalServerError, err, "failed to search content")
		return
	}

	h.sendSuccess(c, http.StatusOK, result, "")
}

// PublishContent publishes content
// @Summary Publish content
// @Tags content
// @Security BearerAuth
// @Param id path string true "Content ID"
// @Success 200 {object} SuccessResponse{data=domain.Content}
// @Failure 403 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Router /api/v1/content/{id}/publish [post]
func (h *Handler) PublishContent(c *gin.Context) {
	userID := middleware.MustGetUserID(c)
	if userID == uuid.Nil {
		return
	}

	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		h.sendError(c, http.StatusBadRequest, err, "invalid content ID")
		return
	}

	content, err := h.contentService.PublishContent(c.Request.Context(), id, userID)
	if err != nil {
		if err == domain.ErrContentNotFound {
			h.sendError(c, http.StatusNotFound, err, "content not found")
			return
		}
		if err == domain.ErrForbidden {
			h.sendError(c, http.StatusForbidden, err, "insufficient permissions")
			return
		}
		h.sendError(c, http.StatusInternalServerError, err, "failed to publish content")
		return
	}

	h.sendSuccess(c, http.StatusOK, content, "content published successfully")
}

// ArchiveContent archives content
// @Summary Archive content
// @Tags content
// @Security BearerAuth
// @Param id path string true "Content ID"
// @Success 200 {object} SuccessResponse{data=domain.Content}
// @Failure 403 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Router /api/v1/content/{id}/archive [post]
func (h *Handler) ArchiveContent(c *gin.Context) {
	userID := middleware.MustGetUserID(c)
	if userID == uuid.Nil {
		return
	}

	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		h.sendError(c, http.StatusBadRequest, err, "invalid content ID")
		return
	}

	content, err := h.contentService.ArchiveContent(c.Request.Context(), id, userID)
	if err != nil {
		if err == domain.ErrContentNotFound {
			h.sendError(c, http.StatusNotFound, err, "content not found")
			return
		}
		if err == domain.ErrForbidden {
			h.sendError(c, http.StatusForbidden, err, "insufficient permissions")
			return
		}
		h.sendError(c, http.StatusInternalServerError, err, "failed to archive content")
		return
	}

	h.sendSuccess(c, http.StatusOK, content, "content archived successfully")
}

// GetContentVersions retrieves version history
// @Summary Get content versions
// @Tags content
// @Produce json
// @Param id path string true "Content ID"
// @Success 200 {object} SuccessResponse{data=[]domain.ContentVersion}
// @Failure 404 {object} ErrorResponse
// @Router /api/v1/content/{id}/versions [get]
func (h *Handler) GetContentVersions(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		h.sendError(c, http.StatusBadRequest, err, "invalid content ID")
		return
	}

	versions, err := h.contentService.GetContentVersions(c.Request.Context(), id)
	if err != nil {
		h.sendError(c, http.StatusInternalServerError, err, "failed to get versions")
		return
	}

	h.sendSuccess(c, http.StatusOK, versions, "")
}

// GetContentVersion retrieves a specific version
// @Summary Get specific content version
// @Tags content
// @Produce json
// @Param id path string true "Content ID"
// @Param version path int true "Version number"
// @Success 200 {object} SuccessResponse{data=domain.ContentVersion}
// @Failure 404 {object} ErrorResponse
// @Router /api/v1/content/{id}/versions/{version} [get]
func (h *Handler) GetContentVersion(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		h.sendError(c, http.StatusBadRequest, err, "invalid content ID")
		return
	}

	var versionNum int
	if err := c.ShouldBindUri(&struct{ Version int `uri:"version" binding:"required"`}{Version: versionNum}); err != nil {
		h.sendError(c, http.StatusBadRequest, err, "invalid version number")
		return
	}

	version, err := h.contentService.GetContentVersion(c.Request.Context(), id, versionNum)
	if err != nil {
		if err == domain.ErrContentNotFound {
			h.sendError(c, http.StatusNotFound, err, "version not found")
			return
		}
		h.sendError(c, http.StatusInternalServerError, err, "failed to get version")
		return
	}

	h.sendSuccess(c, http.StatusOK, version, "")
}
