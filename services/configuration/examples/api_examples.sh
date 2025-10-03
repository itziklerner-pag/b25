#!/bin/bash

# Configuration Service API Examples
# Make sure the service is running on localhost:9096

BASE_URL="http://localhost:9096/api/v1"

echo "=== Configuration Service API Examples ==="
echo ""

# 1. Health Check
echo "1. Health Check"
curl -s http://localhost:9096/health | jq .
echo ""

# 2. Create a Strategy Configuration
echo "2. Create Strategy Configuration"
curl -s -X POST ${BASE_URL}/configurations \
  -H "Content-Type: application/json" \
  -d '{
    "key": "arbitrage_strategy",
    "type": "strategy",
    "value": {
      "name": "Cross-Exchange Arbitrage",
      "type": "arbitrage",
      "enabled": true,
      "parameters": {
        "min_profit_threshold": 0.005,
        "max_position_time": 300,
        "exchanges": ["binance", "coinbase"]
      }
    },
    "format": "json",
    "description": "Cross-exchange arbitrage strategy",
    "created_by": "admin"
  }' | jq .
echo ""

# 3. Create a Risk Limit Configuration
echo "3. Create Risk Limit Configuration"
curl -s -X POST ${BASE_URL}/configurations \
  -H "Content-Type: application/json" \
  -d '{
    "key": "aggressive_risk_limits",
    "type": "risk_limit",
    "value": {
      "max_position_size": 50000,
      "max_loss_per_trade": 2000,
      "max_daily_loss": 10000,
      "max_leverage": 20,
      "stop_loss_percent": 3
    },
    "format": "json",
    "description": "Aggressive risk limits for experienced traders",
    "created_by": "admin"
  }' | jq .
echo ""

# 4. List All Configurations
echo "4. List All Configurations"
curl -s "${BASE_URL}/configurations?limit=10" | jq .
echo ""

# 5. List Strategy Configurations Only
echo "5. List Strategy Configurations"
curl -s "${BASE_URL}/configurations?type=strategy" | jq .
echo ""

# 6. List Active Configurations
echo "6. List Active Configurations"
curl -s "${BASE_URL}/configurations?active=true" | jq .
echo ""

# 7. Get Configuration by Key
echo "7. Get Configuration by Key"
curl -s "${BASE_URL}/configurations/key/default_strategy" | jq .
echo ""

# 8. Update a Configuration (replace ID with actual ID from previous responses)
echo "8. Update Configuration (example - update ID)"
# Get a config ID first
CONFIG_ID=$(curl -s "${BASE_URL}/configurations/key/default_strategy" | jq -r '.data.id')
echo "Updating config: $CONFIG_ID"

curl -s -X PUT ${BASE_URL}/configurations/${CONFIG_ID} \
  -H "Content-Type: application/json" \
  -d '{
    "value": {
      "name": "Market Making",
      "type": "market_making",
      "enabled": true,
      "parameters": {
        "spread": 0.003,
        "order_size": 150
      }
    },
    "format": "json",
    "description": "Updated market making strategy configuration",
    "updated_by": "admin",
    "change_reason": "Increased spread and order size for better profitability"
  }' | jq .
echo ""

# 9. Get Version History
echo "9. Get Version History"
curl -s "${BASE_URL}/configurations/${CONFIG_ID}/versions" | jq .
echo ""

# 10. Get Audit Logs
echo "10. Get Audit Logs"
curl -s "${BASE_URL}/configurations/${CONFIG_ID}/audit-logs?limit=5" | jq .
echo ""

# 11. Deactivate Configuration
echo "11. Deactivate Configuration"
curl -s -X POST ${BASE_URL}/configurations/${CONFIG_ID}/deactivate | jq .
echo ""

# 12. Activate Configuration
echo "12. Activate Configuration"
curl -s -X POST ${BASE_URL}/configurations/${CONFIG_ID}/activate | jq .
echo ""

# 13. Rollback to Previous Version
echo "13. Rollback to Version 1"
curl -s -X POST ${BASE_URL}/configurations/${CONFIG_ID}/rollback \
  -H "Content-Type: application/json" \
  -d '{
    "version": 1,
    "rolled_back_by": "admin",
    "reason": "Reverting changes due to performance issues"
  }' | jq .
echo ""

# 14. Create Trading Pair Configuration
echo "14. Create Trading Pair Configuration"
curl -s -X POST ${BASE_URL}/configurations \
  -H "Content-Type: application/json" \
  -d '{
    "key": "eth_usdt_pair",
    "type": "trading_pair",
    "value": {
      "symbol": "ETH/USDT",
      "base_currency": "ETH",
      "quote_currency": "USDT",
      "min_order_size": 0.01,
      "max_order_size": 100,
      "price_precision": 2,
      "quantity_precision": 8,
      "enabled": true
    },
    "format": "json",
    "description": "Ethereum/USDT trading pair configuration",
    "created_by": "admin"
  }' | jq .
echo ""

# 15. Create System Configuration
echo "15. Create System Configuration"
curl -s -X POST ${BASE_URL}/configurations \
  -H "Content-Type: application/json" \
  -d '{
    "key": "trading_enabled",
    "type": "system",
    "value": {
      "name": "trading_enabled",
      "value": true,
      "type": "boolean"
    },
    "format": "json",
    "description": "Global trading enable/disable flag",
    "created_by": "admin"
  }' | jq .
echo ""

# 16. Delete Configuration (create a test config first)
echo "16. Delete Configuration"
TEST_CONFIG_ID=$(curl -s -X POST ${BASE_URL}/configurations \
  -H "Content-Type: application/json" \
  -d '{
    "key": "test_config_to_delete",
    "type": "system",
    "value": {"test": true},
    "format": "json",
    "created_by": "admin"
  }' | jq -r '.data.id')

echo "Created test config: $TEST_CONFIG_ID"
curl -s -X DELETE ${BASE_URL}/configurations/${TEST_CONFIG_ID} | jq .
echo ""

# 17. Prometheus Metrics
echo "17. Prometheus Metrics (sample)"
curl -s http://localhost:9096/metrics | grep -E "^config_" | head -10
echo ""

echo "=== Examples Complete ==="
