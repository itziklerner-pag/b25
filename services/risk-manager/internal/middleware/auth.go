package middleware

import (
	"context"
	"crypto/subtle"
	"strings"

	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

// AuthInterceptor provides authentication for gRPC calls
type AuthInterceptor struct {
	logger      *zap.Logger
	apiKey      string
	bypassPaths []string
	enabled     bool
}

// NewAuthInterceptor creates a new authentication interceptor
func NewAuthInterceptor(logger *zap.Logger, apiKey string, enabled bool) *AuthInterceptor {
	// Health check and reflection endpoints should bypass auth
	bypassPaths := []string{
		"/grpc.health.v1.Health/Check",
		"/grpc.reflection.v1alpha.ServerReflection/ServerReflectionInfo",
	}

	return &AuthInterceptor{
		logger:      logger,
		apiKey:      apiKey,
		bypassPaths: bypassPaths,
		enabled:     enabled,
	}
}

// Unary returns a server interceptor for unary RPCs
func (i *AuthInterceptor) Unary() grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {
		// Skip authentication if disabled
		if !i.enabled {
			return handler(ctx, req)
		}

		// Check if this is a bypass path
		for _, path := range i.bypassPaths {
			if info.FullMethod == path {
				return handler(ctx, req)
			}
		}

		// Validate API key
		if err := i.validateAPIKey(ctx); err != nil {
			i.logger.Warn("authentication failed",
				zap.String("method", info.FullMethod),
				zap.Error(err),
			)
			return nil, err
		}

		return handler(ctx, req)
	}
}

// Stream returns a server interceptor for streaming RPCs
func (i *AuthInterceptor) Stream() grpc.StreamServerInterceptor {
	return func(
		srv interface{},
		stream grpc.ServerStream,
		info *grpc.StreamServerInfo,
		handler grpc.StreamHandler,
	) error {
		// Skip authentication if disabled
		if !i.enabled {
			return handler(srv, stream)
		}

		// Check if this is a bypass path
		for _, path := range i.bypassPaths {
			if info.FullMethod == path {
				return handler(srv, stream)
			}
		}

		// Validate API key
		if err := i.validateAPIKey(stream.Context()); err != nil {
			i.logger.Warn("authentication failed",
				zap.String("method", info.FullMethod),
				zap.Error(err),
			)
			return err
		}

		return handler(srv, stream)
	}
}

// validateAPIKey validates the API key from the request metadata
func (i *AuthInterceptor) validateAPIKey(ctx context.Context) error {
	// No API key configured - reject all requests (fail secure)
	if i.apiKey == "" {
		return status.Error(codes.Unauthenticated, "authentication not configured")
	}

	// Get metadata from context
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return status.Error(codes.Unauthenticated, "missing metadata")
	}

	// Check for API key in authorization header
	values := md.Get("authorization")
	if len(values) == 0 {
		return status.Error(codes.Unauthenticated, "missing authorization header")
	}

	// Extract token from "Bearer <token>" format
	authHeader := values[0]
	if !strings.HasPrefix(authHeader, "Bearer ") {
		return status.Error(codes.Unauthenticated, "invalid authorization format")
	}

	token := strings.TrimPrefix(authHeader, "Bearer ")

	// Use constant-time comparison to prevent timing attacks
	if subtle.ConstantTimeCompare([]byte(token), []byte(i.apiKey)) != 1 {
		return status.Error(codes.Unauthenticated, "invalid API key")
	}

	return nil
}
