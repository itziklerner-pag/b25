package security

import (
	"fmt"
	"io"
	"net"
	"strings"
	"time"

	"github.com/gabriel-vasile/mimetype"
	"github.com/yourorg/b25/services/media/internal/config"
)

// Scanner defines the interface for virus scanning
type Scanner interface {
	Scan(reader io.Reader) (bool, error)
	IsAvailable() bool
}

// ClamAVScanner implements Scanner using ClamAV
type ClamAVScanner struct {
	host string
	port int
}

// NewClamAVScanner creates a new ClamAV scanner
func NewClamAVScanner(cfg config.ClamAVConfig) (*ClamAVScanner, error) {
	scanner := &ClamAVScanner{
		host: cfg.Host,
		port: cfg.Port,
	}

	// Test connection
	if !scanner.IsAvailable() {
		return nil, fmt.Errorf("ClamAV is not available at %s:%d", cfg.Host, cfg.Port)
	}

	return scanner, nil
}

// Scan scans a file for viruses using ClamAV
func (s *ClamAVScanner) Scan(reader io.Reader) (bool, error) {
	conn, err := net.DialTimeout("tcp", fmt.Sprintf("%s:%d", s.host, s.port), 10*time.Second)
	if err != nil {
		return false, fmt.Errorf("failed to connect to ClamAV: %w", err)
	}
	defer conn.Close()

	// Send INSTREAM command
	if _, err := conn.Write([]byte("zINSTREAM\x00")); err != nil {
		return false, fmt.Errorf("failed to send command: %w", err)
	}

	// Stream file data
	buf := make([]byte, 2048)
	for {
		n, err := reader.Read(buf)
		if n > 0 {
			// Send chunk size (4 bytes, network byte order)
			size := []byte{
				byte(n >> 24),
				byte(n >> 16),
				byte(n >> 8),
				byte(n),
			}
			if _, err := conn.Write(size); err != nil {
				return false, fmt.Errorf("failed to send chunk size: %w", err)
			}

			// Send chunk data
			if _, err := conn.Write(buf[:n]); err != nil {
				return false, fmt.Errorf("failed to send chunk: %w", err)
			}
		}

		if err == io.EOF {
			break
		}
		if err != nil {
			return false, fmt.Errorf("failed to read file: %w", err)
		}
	}

	// Send zero-length chunk to indicate end of stream
	if _, err := conn.Write([]byte{0, 0, 0, 0}); err != nil {
		return false, fmt.Errorf("failed to send end marker: %w", err)
	}

	// Read response
	response := make([]byte, 1024)
	n, err := conn.Read(response)
	if err != nil {
		return false, fmt.Errorf("failed to read response: %w", err)
	}

	result := string(response[:n])

	// Check if file is clean
	if strings.Contains(result, "OK") {
		return true, nil
	}

	// Virus found
	return false, fmt.Errorf("virus detected: %s", result)
}

// IsAvailable checks if ClamAV is available
func (s *ClamAVScanner) IsAvailable() bool {
	conn, err := net.DialTimeout("tcp", fmt.Sprintf("%s:%d", s.host, s.port), 5*time.Second)
	if err != nil {
		return false
	}
	defer conn.Close()

	// Send PING command
	if _, err := conn.Write([]byte("zPING\x00")); err != nil {
		return false
	}

	// Read response
	response := make([]byte, 32)
	n, err := conn.Read(response)
	if err != nil {
		return false
	}

	return strings.Contains(string(response[:n]), "PONG")
}

// NoOpScanner is a no-op scanner for when virus scanning is disabled
type NoOpScanner struct{}

// NewNoOpScanner creates a no-op scanner
func NewNoOpScanner() *NoOpScanner {
	return &NoOpScanner{}
}

// Scan always returns true (file is clean)
func (s *NoOpScanner) Scan(reader io.Reader) (bool, error) {
	return true, nil
}

// IsAvailable always returns true
func (s *NoOpScanner) IsAvailable() bool {
	return true
}

// FileValidator validates file types and sizes
type FileValidator struct {
	allowedMimeTypes []string
	maxFileSize      int64
}

// NewFileValidator creates a new file validator
func NewFileValidator(allowedMimeTypes []string, maxFileSize int64) *FileValidator {
	return &FileValidator{
		allowedMimeTypes: allowedMimeTypes,
		maxFileSize:      maxFileSize,
	}
}

// ValidateFile validates a file's MIME type and size
func (v *FileValidator) ValidateFile(reader io.ReadSeeker, size int64) error {
	// Check file size
	if size > v.maxFileSize {
		return fmt.Errorf("file too large: %d bytes (max: %d)", size, v.maxFileSize)
	}

	// Detect MIME type
	mtype, err := mimetype.DetectReader(reader)
	if err != nil {
		return fmt.Errorf("failed to detect MIME type: %w", err)
	}

	// Reset reader position
	if _, err := reader.Seek(0, 0); err != nil {
		return fmt.Errorf("failed to reset reader: %w", err)
	}

	// Check if MIME type is allowed
	if !v.isAllowedMimeType(mtype.String()) {
		return fmt.Errorf("file type not allowed: %s", mtype.String())
	}

	return nil
}

// isAllowedMimeType checks if a MIME type is in the allowed list
func (v *FileValidator) isAllowedMimeType(mimeType string) bool {
	for _, allowed := range v.allowedMimeTypes {
		if strings.HasPrefix(mimeType, allowed) {
			return true
		}
	}
	return false
}

// SanitizeFileName sanitizes a filename to prevent path traversal attacks
func SanitizeFileName(filename string) string {
	// Remove path separators
	filename = strings.ReplaceAll(filename, "/", "_")
	filename = strings.ReplaceAll(filename, "\\", "_")

	// Remove null bytes
	filename = strings.ReplaceAll(filename, "\x00", "")

	// Remove leading/trailing spaces and dots
	filename = strings.TrimSpace(filename)
	filename = strings.Trim(filename, ".")

	// If filename is empty after sanitization, use a default
	if filename == "" {
		filename = "unnamed"
	}

	return filename
}

// GetFileExtension extracts the file extension from a filename
func GetFileExtension(filename string) string {
	parts := strings.Split(filename, ".")
	if len(parts) > 1 {
		return strings.ToLower(parts[len(parts)-1])
	}
	return ""
}

// IsImageFile checks if a file is an image based on MIME type
func IsImageFile(mimeType string) bool {
	return strings.HasPrefix(mimeType, "image/")
}

// IsVideoFile checks if a file is a video based on MIME type
func IsVideoFile(mimeType string) bool {
	return strings.HasPrefix(mimeType, "video/")
}

// IsAudioFile checks if a file is audio based on MIME type
func IsAudioFile(mimeType string) bool {
	return strings.HasPrefix(mimeType, "audio/")
}
