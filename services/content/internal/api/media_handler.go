package api

import (
	"net/http"

	"github.com/b25/services/content/internal/domain"
	"github.com/b25/services/content/internal/middleware"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// UploadMedia handles media file uploads
// @Summary Upload media file
// @Tags media
// @Security BearerAuth
// @Accept multipart/form-data
// @Produce json
// @Param file formData file true "Media file"
// @Success 201 {object} SuccessResponse{data=domain.Content}
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 413 {object} ErrorResponse
// @Router /api/v1/media/upload [post]
func (h *Handler) UploadMedia(c *gin.Context) {
	userID := middleware.MustGetUserID(c)
	if userID == uuid.Nil {
		return
	}

	file, err := c.FormFile("file")
	if err != nil {
		h.sendError(c, http.StatusBadRequest, err, "file is required")
		return
	}

	content, err := h.mediaService.UploadMedia(c.Request.Context(), file, userID)
	if err != nil {
		if err == domain.ErrMediaTooLarge {
			h.sendError(c, http.StatusRequestEntityTooLarge, err, "file too large")
			return
		}
		if err == domain.ErrInvalidMediaType {
			h.sendError(c, http.StatusBadRequest, err, "invalid media type")
			return
		}
		if err == domain.ErrForbidden {
			h.sendError(c, http.StatusForbidden, err, "insufficient permissions")
			return
		}
		h.logger.Error("failed to upload media", err)
		h.sendError(c, http.StatusInternalServerError, err, "failed to upload media")
		return
	}

	h.sendSuccess(c, http.StatusCreated, content, "media uploaded successfully")
}

// DeleteMedia deletes a media file
// @Summary Delete media
// @Tags media
// @Security BearerAuth
// @Param id path string true "Media content ID"
// @Success 200 {object} SuccessResponse
// @Failure 403 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Router /api/v1/media/{id} [delete]
func (h *Handler) DeleteMedia(c *gin.Context) {
	userID := middleware.MustGetUserID(c)
	if userID == uuid.Nil {
		return
	}

	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		h.sendError(c, http.StatusBadRequest, err, "invalid media ID")
		return
	}

	err = h.mediaService.DeleteMedia(c.Request.Context(), id, userID)
	if err != nil {
		if err == domain.ErrContentNotFound {
			h.sendError(c, http.StatusNotFound, err, "media not found")
			return
		}
		if err == domain.ErrForbidden {
			h.sendError(c, http.StatusForbidden, err, "insufficient permissions")
			return
		}
		h.sendError(c, http.StatusInternalServerError, err, "failed to delete media")
		return
	}

	h.sendSuccess(c, http.StatusOK, nil, "media deleted successfully")
}
