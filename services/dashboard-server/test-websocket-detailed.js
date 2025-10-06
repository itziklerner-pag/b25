#!/usr/bin/env node

// Detailed WebSocket test - shows actual market data
const WebSocket = require('ws');

const WS_URL = 'ws://localhost:8086/ws?type=web&format=json';
const ORIGIN = 'http://localhost:5173';

console.log('Connecting to dashboard-server WebSocket...\n');

const ws = new WebSocket(WS_URL, {
  headers: { 'Origin': ORIGIN }
});

ws.on('open', () => {
  console.log('✓ Connected!\n');
  ws.send(JSON.stringify({
    action: 'subscribe',
    subscriptions: ['market_data']
  }));
});

let updateCount = 0;

ws.on('message', (data) => {
  const msg = JSON.parse(data.toString());
  updateCount++;

  if (msg.type === 'subscribed') {
    console.log(`✓ Subscribed to: ${msg.subscriptions.join(', ')}\n`);
    console.log('Market Data Updates:');
    console.log('-'.repeat(80));
    return;
  }

  if (msg.type === 'snapshot' || msg.type === 'state_update') {
    const marketData = msg.data?.market_data || {};
    const btc = marketData.BTCUSDT;
    const eth = marketData.ETHUSDT;

    if (btc) {
      console.log(`[#${updateCount}] BTC: $${btc.last_price.toFixed(2).padStart(10)} | Bid: $${btc.bid_price.toFixed(2)} | Ask: $${btc.ask_price.toFixed(2)} | Spread: $${(btc.ask_price - btc.bid_price).toFixed(2)}`);
    }
    if (eth) {
      console.log(`      ETH: $${eth.last_price.toFixed(2).padStart(10)} | Bid: $${eth.bid_price.toFixed(2)} | Ask: $${eth.ask_price.toFixed(2)}`);
    }
  }

  if (msg.type === 'diff_update') {
    const changes = msg.changes?.market_data || {};
    Object.keys(changes).forEach(symbol => {
      const data = changes[symbol];
      if (data.last_price) {
        console.log(`[#${updateCount}] ${symbol} price changed to $${data.last_price}`);
      }
    });
  }

  // Stop after 20 updates
  if (updateCount >= 20) {
    console.log('\n✓ Test complete - 20 updates received successfully!');
    ws.close();
  }
});

ws.on('error', (error) => {
  console.error('✗ Error:', error.message);
  process.exit(1);
});

ws.on('close', () => {
  console.log(`\nReceived ${updateCount} updates total`);
  process.exit(0);
});

setTimeout(() => {
  console.log('\n⏱ Timeout - closing connection');
  ws.close();
}, 15000);
