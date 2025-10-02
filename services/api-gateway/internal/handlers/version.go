package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// VersionHandler handles version information
type VersionHandler struct {
	version   string
	buildTime string
	gitCommit string
}

// NewVersionHandler creates a new version handler
func NewVersionHandler(version, buildTime, gitCommit string) *VersionHandler {
	if version == "" {
		version = "1.0.0"
	}
	if buildTime == "" {
		buildTime = "unknown"
	}
	if gitCommit == "" {
		gitCommit = "unknown"
	}

	return &VersionHandler{
		version:   version,
		buildTime: buildTime,
		gitCommit: gitCommit,
	}
}

// Version returns version information
func (v *VersionHandler) Version(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"version":    v.version,
		"build_time": v.buildTime,
		"git_commit": v.gitCommit,
		"service":    "api-gateway",
	})
}
