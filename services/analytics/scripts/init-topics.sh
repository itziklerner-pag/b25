#!/bin/bash
# Initialize Kafka topics for analytics service

set -e

KAFKA_BROKER="${KAFKA_BROKER:-localhost:9092}"

echo "Creating Kafka topics for analytics service..."

# Create topics with replication factor 1 (for development)
# In production, use replication factor 3

kafka-topics --create \
  --bootstrap-server "$KAFKA_BROKER" \
  --topic trading.events \
  --partitions 3 \
  --replication-factor 1 \
  --if-not-exists \
  --config retention.ms=604800000 \
  --config segment.ms=3600000

kafka-topics --create \
  --bootstrap-server "$KAFKA_BROKER" \
  --topic market.data \
  --partitions 5 \
  --replication-factor 1 \
  --if-not-exists \
  --config retention.ms=86400000 \
  --config segment.ms=3600000

kafka-topics --create \
  --bootstrap-server "$KAFKA_BROKER" \
  --topic order.events \
  --partitions 3 \
  --replication-factor 1 \
  --if-not-exists \
  --config retention.ms=604800000 \
  --config segment.ms=3600000

kafka-topics --create \
  --bootstrap-server "$KAFKA_BROKER" \
  --topic user.actions \
  --partitions 2 \
  --replication-factor 1 \
  --if-not-exists \
  --config retention.ms=2592000000 \
  --config segment.ms=86400000

echo "Topics created successfully!"
echo ""
echo "Listing all topics:"
kafka-topics --list --bootstrap-server "$KAFKA_BROKER"
