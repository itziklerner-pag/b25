package processing

import (
	"bytes"
	"fmt"
	"image"
	"image/jpeg"
	"image/png"
	"io"

	"github.com/disintegration/imaging"
	"github.com/yourorg/b25/services/media/internal/config"
	"github.com/yourorg/b25/services/media/internal/models"
	"golang.org/x/image/webp"
)

// ImageProcessor handles image processing operations
type ImageProcessor struct {
	config config.ImageConfig
}

// NewImageProcessor creates a new image processor
func NewImageProcessor(cfg config.ImageConfig) *ImageProcessor {
	return &ImageProcessor{config: cfg}
}

// ProcessImage processes an image and generates thumbnails
func (p *ImageProcessor) ProcessImage(reader io.Reader, originalFormat string) (*ProcessedImage, error) {
	// Decode image
	img, format, err := image.Decode(reader)
	if err != nil {
		return nil, fmt.Errorf("failed to decode image: %w", err)
	}

	// Get original dimensions
	bounds := img.Bounds()
	width := bounds.Dx()
	height := bounds.Dy()

	result := &ProcessedImage{
		OriginalWidth:  width,
		OriginalHeight: height,
		Format:         format,
		Thumbnails:     make(map[string]*Thumbnail),
	}

	// Resize if too large
	if width > p.config.MaxWidth || height > p.config.MaxHeight {
		img = imaging.Fit(img, p.config.MaxWidth, p.config.MaxHeight, imaging.Lanczos)
		bounds = img.Bounds()
		width = bounds.Dx()
		height = bounds.Dy()
	}

	// Encode optimized version
	optimized, err := p.encodeImage(img, format, p.config.Quality)
	if err != nil {
		return nil, fmt.Errorf("failed to encode image: %w", err)
	}
	result.Optimized = optimized
	result.OptimizedWidth = width
	result.OptimizedHeight = height

	// Generate thumbnails
	for _, thumbCfg := range p.config.Thumbnails {
		thumb := imaging.Fit(img, thumbCfg.Width, thumbCfg.Height, imaging.Lanczos)

		encoded, err := p.encodeImage(thumb, format, p.config.Quality)
		if err != nil {
			return nil, fmt.Errorf("failed to encode thumbnail %s: %w", thumbCfg.Name, err)
		}

		thumbBounds := thumb.Bounds()
		result.Thumbnails[thumbCfg.Name] = &Thumbnail{
			Name:   thumbCfg.Name,
			Data:   encoded,
			Width:  thumbBounds.Dx(),
			Height: thumbBounds.Dy(),
		}
	}

	return result, nil
}

// ResizeImage resizes an image to specific dimensions
func (p *ImageProcessor) ResizeImage(reader io.Reader, width, height int) ([]byte, error) {
	img, _, err := image.Decode(reader)
	if err != nil {
		return nil, fmt.Errorf("failed to decode image: %w", err)
	}

	resized := imaging.Fit(img, width, height, imaging.Lanczos)

	var buf bytes.Buffer
	if err := jpeg.Encode(&buf, resized, &jpeg.Options{Quality: p.config.Quality}); err != nil {
		return nil, fmt.Errorf("failed to encode resized image: %w", err)
	}

	return buf.Bytes(), nil
}

// GenerateThumbnail generates a single thumbnail
func (p *ImageProcessor) GenerateThumbnail(reader io.Reader, width, height int) ([]byte, error) {
	img, format, err := image.Decode(reader)
	if err != nil {
		return nil, fmt.Errorf("failed to decode image: %w", err)
	}

	thumb := imaging.Fit(img, width, height, imaging.Lanczos)

	return p.encodeImage(thumb, format, p.config.Quality)
}

// ExtractMetadata extracts metadata from an image
func (p *ImageProcessor) ExtractMetadata(reader io.Reader) (*models.Metadata, error) {
	img, format, err := image.Decode(reader)
	if err != nil {
		return nil, fmt.Errorf("failed to decode image: %w", err)
	}

	bounds := img.Bounds()

	metadata := &models.Metadata{
		Width:  bounds.Dx(),
		Height: bounds.Dy(),
		Format: format,
	}

	return metadata, nil
}

// ConvertFormat converts an image to a different format
func (p *ImageProcessor) ConvertFormat(reader io.Reader, targetFormat string) ([]byte, error) {
	img, _, err := image.Decode(reader)
	if err != nil {
		return nil, fmt.Errorf("failed to decode image: %w", err)
	}

	return p.encodeImage(img, targetFormat, p.config.Quality)
}

// encodeImage encodes an image in the specified format
func (p *ImageProcessor) encodeImage(img image.Image, format string, quality int) ([]byte, error) {
	var buf bytes.Buffer

	switch format {
	case "jpeg", "jpg":
		if err := jpeg.Encode(&buf, img, &jpeg.Options{Quality: quality}); err != nil {
			return nil, err
		}
	case "png":
		encoder := &png.Encoder{CompressionLevel: png.BestCompression}
		if err := encoder.Encode(&buf, img); err != nil {
			return nil, err
		}
	case "webp":
		// Note: webp encoding requires CGO and libwebp
		// For now, fall back to JPEG
		if err := jpeg.Encode(&buf, img, &jpeg.Options{Quality: quality}); err != nil {
			return nil, err
		}
	default:
		return nil, fmt.Errorf("unsupported format: %s", format)
	}

	return buf.Bytes(), nil
}

// Optimize optimizes an image (reduces file size while maintaining quality)
func (p *ImageProcessor) Optimize(reader io.Reader, format string) ([]byte, error) {
	img, _, err := image.Decode(reader)
	if err != nil {
		return nil, fmt.Errorf("failed to decode image: %w", err)
	}

	bounds := img.Bounds()
	width := bounds.Dx()
	height := bounds.Dy()

	// Resize if too large
	if width > p.config.MaxWidth || height > p.config.MaxHeight {
		img = imaging.Fit(img, p.config.MaxWidth, p.config.MaxHeight, imaging.Lanczos)
	}

	return p.encodeImage(img, format, p.config.Quality)
}

// ProcessedImage contains the processed image data
type ProcessedImage struct {
	Optimized       []byte
	OptimizedWidth  int
	OptimizedHeight int
	OriginalWidth   int
	OriginalHeight  int
	Format          string
	Thumbnails      map[string]*Thumbnail
}

// Thumbnail represents a thumbnail image
type Thumbnail struct {
	Name   string
	Data   []byte
	Width  int
	Height int
}

// GetImageDimensions returns the dimensions of an image
func GetImageDimensions(reader io.Reader) (width, height int, err error) {
	config, _, err := image.DecodeConfig(reader)
	if err != nil {
		return 0, 0, fmt.Errorf("failed to decode image config: %w", err)
	}
	return config.Width, config.Height, nil
}

// IsValidImageFormat checks if a format is supported
func IsValidImageFormat(format string) bool {
	validFormats := []string{"jpeg", "jpg", "png", "gif", "webp"}
	for _, f := range validFormats {
		if f == format {
			return true
		}
	}
	return false
}
