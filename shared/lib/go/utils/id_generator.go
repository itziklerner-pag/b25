package utils

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"sync/atomic"
	"time"
)

var (
	sequenceNumber uint64
)

// GenerateOrderID generates a unique order ID with prefix.
func GenerateOrderID(prefix string) string {
	seq := atomic.AddUint64(&sequenceNumber, 1)
	timestamp := time.Now().UnixNano()
	return fmt.Sprintf("%s_%d_%d", prefix, timestamp, seq)
}

// GenerateRequestID generates a unique request ID.
func GenerateRequestID() string {
	return GenerateUUID()
}

// GenerateUUID generates a random UUID-like string.
func GenerateUUID() string {
	b := make([]byte, 16)
	_, err := rand.Read(b)
	if err != nil {
		// Fallback to time-based ID if random fails
		return fmt.Sprintf("%d_%d", time.Now().UnixNano(), atomic.AddUint64(&sequenceNumber, 1))
	}
	return fmt.Sprintf("%x-%x-%x-%x-%x", b[0:4], b[4:6], b[6:8], b[8:10], b[10:])
}

// GenerateClientOrderID generates a client order ID for exchange submission.
func GenerateClientOrderID(strategyID string) string {
	timestamp := time.Now().UnixNano() / 1e6 // milliseconds
	seq := atomic.AddUint64(&sequenceNumber, 1)
	return fmt.Sprintf("%s_%d_%d", strategyID, timestamp, seq)
}

// GenerateTraceID generates a trace ID for distributed tracing.
func GenerateTraceID() string {
	b := make([]byte, 16)
	_, err := rand.Read(b)
	if err != nil {
		return fmt.Sprintf("%d", time.Now().UnixNano())
	}
	return base64.RawURLEncoding.EncodeToString(b)
}

// GenerateSpanID generates a span ID for distributed tracing.
func GenerateSpanID() string {
	b := make([]byte, 8)
	_, err := rand.Read(b)
	if err != nil {
		return fmt.Sprintf("%d", atomic.AddUint64(&sequenceNumber, 1))
	}
	return base64.RawURLEncoding.EncodeToString(b)
}
