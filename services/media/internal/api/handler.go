package api

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"path/filepath"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	log "github.com/sirupsen/logrus"

	"github.com/yourorg/b25/services/media/internal/database"
	"github.com/yourorg/b25/services/media/internal/models"
	"github.com/yourorg/b25/services/media/internal/security"
)

// Handler handles HTTP requests
type Handler struct {
	config ServerConfig
	repo   *database.MediaRepository
}

// NewHandler creates a new request handler
func NewHandler(cfg ServerConfig) *Handler {
	return &Handler{
		config: cfg,
		repo:   database.NewMediaRepository(cfg.DB),
	}
}

// Health returns the health status
func (h *Handler) Health(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status": "healthy",
		"time":   time.Now().Format(time.RFC3339),
	})
}

// Metrics returns Prometheus metrics
func (h *Handler) Metrics(c *gin.Context) {
	promhttp.Handler().ServeHTTP(c.Writer, c.Request)
}

// UploadMedia handles single file upload
func (h *Handler) UploadMedia(c *gin.Context) {
	// Get user/org IDs from headers
	userID := c.GetHeader("X-User-ID")
	if userID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "X-User-ID header is required"})
		return
	}

	orgID := c.GetHeader("X-Org-ID")

	// Parse multipart form
	file, header, err := c.Request.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to get file from request"})
		return
	}
	defer file.Close()

	// Check quota
	hasQuota, err := h.config.QuotaManager.CheckQuota(userID, "user", header.Size)
	if err != nil {
		log.Errorf("Failed to check quota: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to check quota"})
		return
	}

	if !hasQuota {
		c.JSON(http.StatusForbidden, gin.H{"error": "Quota exceeded"})
		return
	}

	// Validate file
	validator := security.NewFileValidator(
		h.config.Config.Security.AllowedMimeTypes,
		h.config.Config.Security.MaxFileSize,
	)

	if err := validator.ValidateFile(file, header.Size); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Reset file position after validation
	file.Seek(0, 0)

	// Scan for viruses if enabled
	if h.config.Config.Security.EnableVirusScan {
		clean, err := h.config.SecurityScanner.Scan(file)
		if err != nil || !clean {
			c.JSON(http.StatusBadRequest, gin.H{"error": "File failed security scan"})
			return
		}
		file.Seek(0, 0)
	}

	// Detect MIME type
	mimeType := header.Header.Get("Content-Type")

	// Generate unique filename
	mediaID := uuid.New().String()
	extension := filepath.Ext(header.Filename)
	fileName := fmt.Sprintf("%s%s", mediaID, extension)

	// Determine media type
	var mediaType models.MediaType
	if security.IsImageFile(mimeType) {
		mediaType = models.MediaTypeImage
	} else if security.IsVideoFile(mimeType) {
		mediaType = models.MediaTypeVideo
	} else if security.IsAudioFile(mimeType) {
		mediaType = models.MediaTypeAudio
	} else {
		mediaType = models.MediaTypeFile
	}

	// Create storage path
	storagePath := fmt.Sprintf("%s/%s/%s", userID, time.Now().Format("2006/01/02"), fileName)

	// Upload to storage
	_, err = h.config.Storage.Upload(storagePath, file, mimeType)
	if err != nil {
		log.Errorf("Failed to upload file: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to upload file"})
		return
	}

	// Get public URL
	publicURL := h.config.Storage.GetURL(storagePath)

	// Create media record
	media := &models.Media{
		ID:           mediaID,
		UserID:       userID,
		OrgID:        orgID,
		FileName:     fileName,
		OriginalName: security.SanitizeFileName(header.Filename),
		MimeType:     mimeType,
		MediaType:    mediaType,
		Size:         header.Size,
		Status:       models.MediaStatusPending,
		StoragePath:  storagePath,
		PublicURL:    publicURL,
		Metadata:     models.Metadata{},
	}

	// Save to database
	if err := h.repo.Create(media); err != nil {
		log.Errorf("Failed to create media record: %v", err)
		// Cleanup uploaded file
		h.config.Storage.Delete(storagePath)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create media record"})
		return
	}

	// Update quota
	if err := h.config.QuotaManager.UpdateQuota(userID, "user", header.Size); err != nil {
		log.Errorf("Failed to update quota: %v", err)
	}

	// Process media asynchronously
	go h.processMedia(media)

	c.JSON(http.StatusCreated, media)
}

// processMedia processes uploaded media (thumbnails, transcoding, etc.)
func (h *Handler) processMedia(media *models.Media) {
	// Update status to processing
	h.repo.UpdateStatus(media.ID, models.MediaStatusProcessing)

	// Download file from storage
	reader, err := h.config.Storage.Download(media.StoragePath)
	if err != nil {
		log.Errorf("Failed to download media for processing: %v", err)
		h.repo.UpdateStatus(media.ID, models.MediaStatusFailed)
		return
	}
	defer reader.Close()

	// Process based on media type
	switch media.MediaType {
	case models.MediaTypeImage:
		h.processImage(media, reader)
	case models.MediaTypeVideo:
		h.processVideo(media, reader)
	default:
		// No processing needed
		h.repo.UpdateStatus(media.ID, models.MediaStatusReady)
	}
}

// processImage processes an image
func (h *Handler) processImage(media *models.Media, reader io.ReadCloser) {
	// Read all data
	data, err := io.ReadAll(reader)
	if err != nil {
		log.Errorf("Failed to read image data: %v", err)
		h.repo.UpdateStatus(media.ID, models.MediaStatusFailed)
		return
	}

	// Process image
	processed, err := h.config.ImageProcessor.ProcessImage(bytes.NewReader(data), security.GetFileExtension(media.FileName))
	if err != nil {
		log.Errorf("Failed to process image: %v", err)
		h.repo.UpdateStatus(media.ID, models.MediaStatusFailed)
		return
	}

	// Upload thumbnails
	thumbnailInfos := []models.ThumbnailInfo{}
	for name, thumb := range processed.Thumbnails {
		thumbPath := fmt.Sprintf("%s/thumbnails/%s_%s", filepath.Dir(media.StoragePath), name, media.FileName)
		_, err := h.config.Storage.Upload(thumbPath, bytes.NewReader(thumb.Data), media.MimeType)
		if err != nil {
			log.Errorf("Failed to upload thumbnail: %v", err)
			continue
		}

		thumbnailInfos = append(thumbnailInfos, models.ThumbnailInfo{
			Name:   name,
			URL:    h.config.Storage.GetURL(thumbPath),
			Width:  thumb.Width,
			Height: thumb.Height,
		})
	}

	// Update metadata
	now := time.Now()
	metadata := models.Metadata{
		Width:       processed.OptimizedWidth,
		Height:      processed.OptimizedHeight,
		Format:      processed.Format,
		Thumbnails:  thumbnailInfos,
		ProcessedAt: &now,
	}

	h.repo.UpdateMetadata(media.ID, metadata)
	h.repo.UpdateStatus(media.ID, models.MediaStatusReady)
}

// processVideo processes a video (placeholder - would need actual implementation)
func (h *Handler) processVideo(media *models.Media, reader io.ReadCloser) {
	// For actual implementation, you would:
	// 1. Save the video to a temporary file
	// 2. Use VideoProcessor to transcode to different qualities
	// 3. Generate thumbnail
	// 4. Upload all variants and thumbnail
	// 5. Update metadata

	// For now, just mark as ready
	h.repo.UpdateStatus(media.ID, models.MediaStatusReady)
}

// GetMedia retrieves media by ID
func (h *Handler) GetMedia(c *gin.Context) {
	id := c.Param("id")

	media, err := h.repo.GetByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Media not found"})
		return
	}

	c.JSON(http.StatusOK, media)
}

// DownloadMedia downloads media file
func (h *Handler) DownloadMedia(c *gin.Context) {
	id := c.Param("id")

	media, err := h.repo.GetByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Media not found"})
		return
	}

	// Download from storage
	reader, err := h.config.Storage.Download(media.StoragePath)
	if err != nil {
		log.Errorf("Failed to download media: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to download media"})
		return
	}
	defer reader.Close()

	// Set headers
	c.Header("Content-Type", media.MimeType)
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s", media.OriginalName))

	// Stream file
	io.Copy(c.Writer, reader)
}

// StreamMedia streams media file with range support
func (h *Handler) StreamMedia(c *gin.Context) {
	id := c.Param("id")

	media, err := h.repo.GetByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Media not found"})
		return
	}

	// Download from storage
	reader, err := h.config.Storage.Download(media.StoragePath)
	if err != nil {
		log.Errorf("Failed to download media: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to download media"})
		return
	}
	defer reader.Close()

	// Set headers for streaming
	c.Header("Content-Type", media.MimeType)
	c.Header("Accept-Ranges", "bytes")
	c.Header("Content-Length", strconv.FormatInt(media.Size, 10))

	// For proper range request support, you would need to implement
	// HTTP range parsing and seeking. For now, stream the entire file.
	io.Copy(c.Writer, reader)
}

// ListMedia lists media with filters
func (h *Handler) ListMedia(c *gin.Context) {
	filter := models.MediaFilter{
		UserID:    c.Query("user_id"),
		OrgID:     c.Query("org_id"),
		Limit:     50,
		Offset:    0,
		SortBy:    "created_at",
		SortOrder: "desc",
	}

	if mediaType := c.Query("type"); mediaType != "" {
		filter.MediaType = models.MediaType(mediaType)
	}

	if status := c.Query("status"); status != "" {
		filter.Status = models.MediaStatus(status)
	}

	if limit := c.Query("limit"); limit != "" {
		if l, err := strconv.Atoi(limit); err == nil {
			filter.Limit = l
		}
	}

	if offset := c.Query("offset"); offset != "" {
		if o, err := strconv.Atoi(offset); err == nil {
			filter.Offset = o
		}
	}

	mediaList, err := h.repo.List(filter)
	if err != nil {
		log.Errorf("Failed to list media: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list media"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"media":  mediaList,
		"count":  len(mediaList),
		"limit":  filter.Limit,
		"offset": filter.Offset,
	})
}

// DeleteMedia deletes media
func (h *Handler) DeleteMedia(c *gin.Context) {
	id := c.Param("id")

	media, err := h.repo.GetByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Media not found"})
		return
	}

	// Delete from storage
	if err := h.config.Storage.Delete(media.StoragePath); err != nil {
		log.Errorf("Failed to delete from storage: %v", err)
	}

	// Soft delete from database
	if err := h.repo.Delete(id); err != nil {
		log.Errorf("Failed to delete media: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete media"})
		return
	}

	// Update quota
	if err := h.config.QuotaManager.UpdateQuota(media.UserID, "user", -media.Size); err != nil {
		log.Errorf("Failed to update quota: %v", err)
	}

	c.JSON(http.StatusOK, gin.H{"message": "Media deleted successfully"})
}

// GetMetadata retrieves media metadata
func (h *Handler) GetMetadata(c *gin.Context) {
	id := c.Param("id")

	media, err := h.repo.GetByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Media not found"})
		return
	}

	c.JSON(http.StatusOK, media.Metadata)
}

// GetThumbnail retrieves a thumbnail
func (h *Handler) GetThumbnail(c *gin.Context) {
	id := c.Param("id")
	size := c.Param("size")

	media, err := h.repo.GetByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Media not found"})
		return
	}

	// Find thumbnail in metadata
	var thumbnailURL string
	for _, thumb := range media.Metadata.Thumbnails {
		if thumb.Name == size {
			thumbnailURL = thumb.URL
			break
		}
	}

	if thumbnailURL == "" {
		c.JSON(http.StatusNotFound, gin.H{"error": "Thumbnail not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"url": thumbnailURL})
}

// GetUserQuota retrieves user quota
func (h *Handler) GetUserQuota(c *gin.Context) {
	userID := c.Param("user_id")

	quota, err := h.config.QuotaManager.GetQuota(userID, "user")
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Quota not found"})
		return
	}

	c.JSON(http.StatusOK, quota)
}

// GetOrgQuota retrieves organization quota
func (h *Handler) GetOrgQuota(c *gin.Context) {
	orgID := c.Param("org_id")

	quota, err := h.config.QuotaManager.GetQuota(orgID, "org")
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Quota not found"})
		return
	}

	c.JSON(http.StatusOK, quota)
}

// SetUserQuota updates user quota limit
func (h *Handler) SetUserQuota(c *gin.Context) {
	userID := c.Param("user_id")

	var req struct {
		Limit int64 `json:"limit"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	if err := h.config.QuotaManager.SetQuotaLimit(userID, "user", req.Limit); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update quota"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Quota updated successfully"})
}

// SetOrgQuota updates organization quota limit
func (h *Handler) SetOrgQuota(c *gin.Context) {
	orgID := c.Param("org_id")

	var req struct {
		Limit int64 `json:"limit"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	if err := h.config.QuotaManager.SetQuotaLimit(orgID, "org", req.Limit); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update quota"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Quota updated successfully"})
}

// GetUserStats retrieves user statistics
func (h *Handler) GetUserStats(c *gin.Context) {
	userID := c.Param("user_id")

	stats, err := h.repo.GetStats(userID, "user")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get stats"})
		return
	}

	c.JSON(http.StatusOK, stats)
}

// GetOrgStats retrieves organization statistics
func (h *Handler) GetOrgStats(c *gin.Context) {
	orgID := c.Param("org_id")

	stats, err := h.repo.GetStats(orgID, "org")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get stats"})
		return
	}

	c.JSON(http.StatusOK, stats)
}

// RecalculateQuota recalculates quota for an entity
func (h *Handler) RecalculateQuota(c *gin.Context) {
	entityID := c.Param("entity_id")
	entityType := c.Query("type") // "user" or "org"

	if entityType != "user" && entityType != "org" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid entity type"})
		return
	}

	if err := h.config.QuotaManager.RecalculateQuota(entityID, entityType); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to recalculate quota"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Quota recalculated successfully"})
}

// GetStorageStatus returns storage system status
func (h *Handler) GetStorageStatus(c *gin.Context) {
	// This is a placeholder - actual implementation would check storage backend health
	c.JSON(http.StatusOK, gin.H{
		"status": "healthy",
		"type":   h.config.Config.Storage.Type,
	})
}

// Placeholder handlers for multipart upload
func (h *Handler) InitMultipartUpload(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{"error": "Multipart upload not yet implemented"})
}

func (h *Handler) UploadChunk(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{"error": "Multipart upload not yet implemented"})
}

func (h *Handler) CompleteMultipartUpload(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{"error": "Multipart upload not yet implemented"})
}

func (h *Handler) AbortMultipartUpload(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{"error": "Multipart upload not yet implemented"})
}

func (h *Handler) ResizeImage(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{"error": "Image resize not yet implemented"})
}

func (h *Handler) ConvertFormat(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{"error": "Format conversion not yet implemented"})
}

func (h *Handler) GetVideoVariants(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{"error": "Video variants not yet implemented"})
}

func (h *Handler) GetHLSPlaylist(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{"error": "HLS playlist not yet implemented"})
}
