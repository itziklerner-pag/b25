package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/yourusername/b25/tests/testutil/exchange"
)

func main() {
	// Parse configuration from environment
	config := &exchange.MockExchangeConfig{
		HTTPAddr:           getEnv("HTTP_ADDR", ":8545"),
		WSAddr:             getEnv("WS_ADDR", ":8546"),
		OrderLatency:       parseDuration(getEnv("ORDER_LATENCY", "10ms")),
		FillDelay:          parseDuration(getEnv("FILL_DELAY", "50ms")),
		RejectRate:         parseFloat(getEnv("REJECT_RATE", "0.0")),
		PartialFillEnabled: parseBool(getEnv("PARTIAL_FILL_ENABLED", "false")),
		MarketDataEnabled:  parseBool(getEnv("MARKET_DATA_ENABLED", "true")),
	}

	// Create mock exchange
	mockExchange := exchange.NewMockExchange(config)

	// Start mock exchange
	if err := mockExchange.Start(); err != nil {
		log.Fatalf("Failed to start mock exchange: %v", err)
	}

	log.Printf("Mock Exchange started")
	log.Printf("HTTP API listening on %s", config.HTTPAddr)
	log.Printf("WebSocket listening on %s", config.WSAddr)

	// Wait for shutdown signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	log.Println("Shutting down mock exchange...")
	if err := mockExchange.Stop(); err != nil {
		log.Printf("Error stopping mock exchange: %v", err)
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func parseDuration(s string) time.Duration {
	d, err := time.ParseDuration(s)
	if err != nil {
		return 10 * time.Millisecond
	}
	return d
}

func parseFloat(s string) float64 {
	var f float64
	fmt.Sscanf(s, "%f", &f)
	return f
}

func parseBool(s string) bool {
	return s == "true" || s == "1" || s == "yes"
}
