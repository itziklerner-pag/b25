package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/nats-io/nats.go"
)

// ConfigUpdateEvent represents a configuration update event
type ConfigUpdateEvent struct {
	ID        string          `json:"id"`
	Key       string          `json:"key"`
	Type      string          `json:"type"`
	Value     json.RawMessage `json:"value"`
	Format    string          `json:"format"`
	Version   int             `json:"version"`
	Action    string          `json:"action"`
	Timestamp time.Time       `json:"timestamp"`
}

func main() {
	// Connect to NATS
	natsURL := os.Getenv("NATS_URL")
	if natsURL == "" {
		natsURL = "nats://localhost:4222"
	}

	nc, err := nats.Connect(natsURL)
	if err != nil {
		log.Fatal(err)
	}
	defer nc.Close()

	fmt.Printf("Connected to NATS at %s\n", natsURL)
	fmt.Println("Subscribing to configuration updates...")
	fmt.Println("Topic pattern: config.updates.*")
	fmt.Println("")

	// Subscribe to all configuration updates
	_, err = nc.Subscribe("config.updates.*", func(msg *nats.Msg) {
		var event ConfigUpdateEvent
		if err := json.Unmarshal(msg.Data, &event); err != nil {
			log.Printf("Error unmarshaling event: %v\n", err)
			return
		}

		fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
		fmt.Printf("🔔 Configuration Update Received\n")
		fmt.Printf("   Topic:     %s\n", msg.Subject)
		fmt.Printf("   ID:        %s\n", event.ID)
		fmt.Printf("   Key:       %s\n", event.Key)
		fmt.Printf("   Type:      %s\n", event.Type)
		fmt.Printf("   Action:    %s\n", event.Action)
		fmt.Printf("   Version:   %d\n", event.Version)
		fmt.Printf("   Format:    %s\n", event.Format)
		fmt.Printf("   Timestamp: %s\n", event.Timestamp.Format(time.RFC3339))
		fmt.Printf("   Value:     %s\n", string(event.Value))
		fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
		fmt.Println("")

		// Here you would implement the actual hot-reload logic
		// For example:
		// - Update in-memory configuration cache
		// - Reload strategy parameters
		// - Adjust risk limits
		// - Enable/disable trading pairs
		handleConfigUpdate(&event)
	})

	if err != nil {
		log.Fatal(err)
	}

	// Subscribe to specific config types
	subscribeToConfigType(nc, "strategy")
	subscribeToConfigType(nc, "risk_limit")
	subscribeToConfigType(nc, "trading_pair")
	subscribeToConfigType(nc, "system")

	fmt.Println("Listening for configuration updates... Press Ctrl+C to exit")

	// Wait for interrupt signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	fmt.Println("\nShutting down...")
}

func subscribeToConfigType(nc *nats.Conn, configType string) {
	subject := fmt.Sprintf("config.updates.%s", configType)
	_, err := nc.Subscribe(subject, func(msg *nats.Msg) {
		var event ConfigUpdateEvent
		if err := json.Unmarshal(msg.Data, &event); err != nil {
			log.Printf("Error unmarshaling %s event: %v\n", configType, err)
			return
		}

		fmt.Printf("📋 [%s] %s: %s (v%d)\n", configType, event.Action, event.Key, event.Version)
	})

	if err != nil {
		log.Printf("Error subscribing to %s: %v\n", subject, err)
	}
}

func handleConfigUpdate(event *ConfigUpdateEvent) {
	// This is where you implement hot-reload logic based on config type
	switch event.Type {
	case "strategy":
		fmt.Println("   → Reloading strategy configuration...")
		// Update strategy parameters in memory
		// Restart strategy if needed

	case "risk_limit":
		fmt.Println("   → Updating risk limits...")
		// Update risk limits in risk manager
		// May need to adjust current positions

	case "trading_pair":
		fmt.Println("   → Updating trading pair settings...")
		// Enable/disable trading pair
		// Update order size limits

	case "system":
		fmt.Println("   → Applying system configuration...")
		// Update system-level settings
		// May trigger service-wide changes

	default:
		fmt.Printf("   → Unknown config type: %s\n", event.Type)
	}
}
