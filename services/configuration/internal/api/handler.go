package api

import (
	"database/sql"

	"github.com/b25/services/configuration/internal/service"
	"github.com/nats-io/nats.go"
	"go.uber.org/zap"
)

// Handler contains all API handlers
type Handler struct {
	configService *service.ConfigurationService
	logger        *zap.Logger
	db            *sql.DB
	natsConn      *nats.Conn
}

// NewHandler creates a new API handler
func NewHandler(configService *service.ConfigurationService, logger *zap.Logger, db *sql.DB, natsConn *nats.Conn) *Handler {
	return &Handler{
		configService: configService,
		logger:        logger,
		db:            db,
		natsConn:      natsConn,
	}
}

// Response represents a standard API response
type Response struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
	Message string      `json:"message,omitempty"`
}

// PaginatedResponse represents a paginated API response
type PaginatedResponse struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data"`
	Total   int         `json:"total"`
	Limit   int         `json:"limit"`
	Offset  int         `json:"offset"`
	Error   string      `json:"error,omitempty"`
}

// ErrorResponse represents an error response
type ErrorResponse struct {
	Success bool   `json:"success"`
	Error   string `json:"error"`
	Code    string `json:"code,omitempty"`
}
