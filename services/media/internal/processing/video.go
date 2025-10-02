package processing

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/yourorg/b25/services/media/internal/config"
	"github.com/yourorg/b25/services/media/internal/models"
)

// VideoProcessor handles video processing operations
type VideoProcessor struct {
	config config.VideoConfig
}

// NewVideoProcessor creates a new video processor
func NewVideoProcessor(cfg config.VideoConfig) *VideoProcessor {
	return &VideoProcessor{config: cfg}
}

// ProcessVideo transcodes a video to multiple profiles
func (p *VideoProcessor) ProcessVideo(inputPath string, outputDir string) (*ProcessedVideo, error) {
	// Check if FFmpeg is available
	if !p.isFFmpegAvailable() {
		return nil, fmt.Errorf("FFmpeg is not installed or not available in PATH")
	}

	result := &ProcessedVideo{
		Variants: make(map[string]*VideoVariant),
	}

	// Get video metadata
	metadata, err := p.GetMetadata(inputPath)
	if err != nil {
		return nil, fmt.Errorf("failed to get video metadata: %w", err)
	}
	result.Metadata = metadata

	// Ensure output directory exists
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create output directory: %w", err)
	}

	// Process each profile
	for _, profile := range p.config.Profiles {
		outputPath := filepath.Join(outputDir, fmt.Sprintf("%s.mp4", profile.Name))

		if err := p.transcodeVideo(inputPath, outputPath, profile); err != nil {
			return nil, fmt.Errorf("failed to transcode to %s: %w", profile.Name, err)
		}

		// Get file size
		info, err := os.Stat(outputPath)
		if err != nil {
			return nil, fmt.Errorf("failed to stat output file: %w", err)
		}

		result.Variants[profile.Name] = &VideoVariant{
			Name:    profile.Name,
			Path:    outputPath,
			Width:   profile.Width,
			Height:  profile.Height,
			Bitrate: profile.Bitrate,
			Size:    info.Size(),
		}
	}

	return result, nil
}

// GenerateThumbnail generates a thumbnail from a video
func (p *VideoProcessor) GenerateThumbnail(inputPath, outputPath string) error {
	if !p.isFFmpegAvailable() {
		return fmt.Errorf("FFmpeg is not installed")
	}

	// Ensure output directory exists
	outputDir := filepath.Dir(outputPath)
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	cmd := exec.Command(
		"ffmpeg",
		"-i", inputPath,
		"-ss", p.config.Thumbnail.Time,
		"-vframes", "1",
		"-vf", fmt.Sprintf("scale=%d:%d", p.config.Thumbnail.Width, p.config.Thumbnail.Height),
		"-y",
		outputPath,
	)

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("ffmpeg thumbnail generation failed: %s - %w", string(output), err)
	}

	return nil
}

// GetMetadata extracts metadata from a video file
func (p *VideoProcessor) GetMetadata(inputPath string) (*models.Metadata, error) {
	if !p.isFFmpegAvailable() {
		return nil, fmt.Errorf("FFmpeg is not installed")
	}

	// Use ffprobe to get video metadata
	cmd := exec.Command(
		"ffprobe",
		"-v", "error",
		"-select_streams", "v:0",
		"-show_entries", "stream=width,height,duration,bit_rate",
		"-of", "default=noprint_wrappers=1",
		inputPath,
	)

	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("ffprobe failed: %w", err)
	}

	metadata := &models.Metadata{}

	// Parse output
	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		parts := strings.Split(line, "=")
		if len(parts) != 2 {
			continue
		}

		key := parts[0]
		value := parts[1]

		switch key {
		case "width":
			if w, err := strconv.Atoi(value); err == nil {
				metadata.Width = w
			}
		case "height":
			if h, err := strconv.Atoi(value); err == nil {
				metadata.Height = h
			}
		case "duration":
			if d, err := strconv.ParseFloat(value, 64); err == nil {
				metadata.Duration = d
			}
		case "bit_rate":
			if b, err := strconv.Atoi(value); err == nil {
				metadata.Bitrate = b
			}
		}
	}

	return metadata, nil
}

// transcodeVideo transcodes a video to a specific profile
func (p *VideoProcessor) transcodeVideo(inputPath, outputPath string, profile config.ProfileConfig) error {
	cmd := exec.Command(
		"ffmpeg",
		"-i", inputPath,
		"-c:v", p.config.Codecs.Video,
		"-b:v", profile.Bitrate,
		"-vf", fmt.Sprintf("scale=%d:%d", profile.Width, profile.Height),
		"-c:a", p.config.Codecs.Audio,
		"-b:a", "128k",
		"-movflags", "+faststart", // Enable streaming
		"-y",
		outputPath,
	)

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("ffmpeg transcode failed: %s - %w", string(output), err)
	}

	return nil
}

// ExtractAudio extracts audio from a video
func (p *VideoProcessor) ExtractAudio(inputPath, outputPath string) error {
	if !p.isFFmpegAvailable() {
		return fmt.Errorf("FFmpeg is not installed")
	}

	cmd := exec.Command(
		"ffmpeg",
		"-i", inputPath,
		"-vn",
		"-acodec", "libmp3lame",
		"-y",
		outputPath,
	)

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("ffmpeg audio extraction failed: %s - %w", string(output), err)
	}

	return nil
}

// CreateHLSPlaylist creates an HLS playlist for adaptive streaming
func (p *VideoProcessor) CreateHLSPlaylist(inputPath, outputDir string) error {
	if !p.isFFmpegAvailable() {
		return fmt.Errorf("FFmpeg is not installed")
	}

	playlistPath := filepath.Join(outputDir, "playlist.m3u8")

	cmd := exec.Command(
		"ffmpeg",
		"-i", inputPath,
		"-c:v", p.config.Codecs.Video,
		"-c:a", p.config.Codecs.Audio,
		"-hls_time", "10",
		"-hls_playlist_type", "vod",
		"-hls_segment_filename", filepath.Join(outputDir, "segment_%03d.ts"),
		"-y",
		playlistPath,
	)

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("ffmpeg HLS creation failed: %s - %w", string(output), err)
	}

	return nil
}

// isFFmpegAvailable checks if FFmpeg is available in the system
func (p *VideoProcessor) isFFmpegAvailable() bool {
	_, err := exec.LookPath("ffmpeg")
	return err == nil
}

// ValidateVideo validates a video file
func (p *VideoProcessor) ValidateVideo(inputPath string) error {
	// Check file size
	info, err := os.Stat(inputPath)
	if err != nil {
		return fmt.Errorf("failed to stat video file: %w", err)
	}

	if info.Size() > p.config.MaxSize {
		return fmt.Errorf("video file too large: %d bytes (max: %d)", info.Size(), p.config.MaxSize)
	}

	// Get metadata to validate duration
	metadata, err := p.GetMetadata(inputPath)
	if err != nil {
		return fmt.Errorf("failed to get video metadata: %w", err)
	}

	if metadata.Duration > float64(p.config.MaxDuration) {
		return fmt.Errorf("video too long: %.2f seconds (max: %d)", metadata.Duration, p.config.MaxDuration)
	}

	return nil
}

// ProcessedVideo contains processed video data
type ProcessedVideo struct {
	Metadata *models.Metadata
	Variants map[string]*VideoVariant
}

// VideoVariant represents a transcoded video variant
type VideoVariant struct {
	Name    string
	Path    string
	Width   int
	Height  int
	Bitrate string
	Size    int64
}
