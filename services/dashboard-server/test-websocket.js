#!/usr/bin/env node

// Test WebSocket client for dashboard-server
// Tests: Connection, subscription, origin checking, and data flow

const WebSocket = require('ws');

console.log('='.repeat(50));
console.log('Dashboard Server WebSocket Test');
console.log('='.repeat(50));
console.log('');

// Test configuration
const WS_URL = 'ws://localhost:8086/ws?type=web&format=json';
const ORIGIN = 'http://localhost:5173'; // Should be allowed

console.log(`Connecting to: ${WS_URL}`);
console.log(`Origin: ${ORIGIN}`);
console.log('');

// Create WebSocket with proper origin header
const ws = new WebSocket(WS_URL, {
  headers: {
    'Origin': ORIGIN
  }
});

let messageCount = 0;
let startTime = Date.now();

ws.on('open', () => {
  console.log('✓ Connected successfully!');
  console.log('');

  // Subscribe to market data
  console.log('Subscribing to market_data...');
  ws.send(JSON.stringify({
    action: 'subscribe',
    subscriptions: ['market_data']
  }));
});

ws.on('message', (data) => {
  messageCount++;
  const elapsed = ((Date.now() - startTime) / 1000).toFixed(1);

  try {
    const msg = JSON.parse(data.toString());

    if (msg.type === 'subscribed') {
      console.log(`✓ Subscription confirmed!`);
      console.log(`  Subscriptions: ${msg.subscriptions.join(', ')}`);
      console.log(`  Sequence: ${msg.sequence}`);
      console.log('');
      console.log('Waiting for updates... (Ctrl+C to stop)');
      console.log('-'.repeat(50));
    } else if (msg.type === 'state_update') {
      console.log(`[${elapsed}s] State Update (#${messageCount})`);
      console.log(`  Sequence: ${msg.sequence}`);
      if (msg.data.market_data) {
        const symbols = Object.keys(msg.data.market_data);
        console.log(`  Symbols: ${symbols.length} (${symbols.join(', ')})`);

        if (msg.data.market_data.BTCUSDT) {
          const btc = msg.data.market_data.BTCUSDT;
          console.log(`  BTC: $${btc.last_price} (bid: $${btc.bid_price}, ask: $${btc.ask_price})`);
        }
      }
      console.log('');
    } else if (msg.type === 'diff_update') {
      console.log(`[${elapsed}s] Diff Update (#${messageCount})`);
      console.log(`  Sequence: ${msg.sequence}`);
      if (msg.changes?.market_data) {
        const changed = Object.keys(msg.changes.market_data);
        console.log(`  Changed symbols: ${changed.join(', ')}`);

        if (msg.changes.market_data.BTCUSDT) {
          const btc = msg.changes.market_data.BTCUSDT;
          console.log(`  BTC: $${btc.last_price || 'no change'}`);
        }
      }
      console.log('');
    } else {
      console.log(`[${elapsed}s] ${msg.type} (#${messageCount})`);
      console.log(`  ${JSON.stringify(msg).substring(0, 100)}...`);
      console.log('');
    }
  } catch (error) {
    console.error('Parse error:', error.message);
  }
});

ws.on('error', (error) => {
  console.error('✗ WebSocket error:', error.message);
});

ws.on('close', (code, reason) => {
  const elapsed = ((Date.now() - startTime) / 1000).toFixed(1);
  console.log('');
  console.log('='.repeat(50));
  console.log(`Connection closed after ${elapsed}s`);
  console.log(`Code: ${code}`);
  console.log(`Reason: ${reason || 'Normal closure'}`);
  console.log(`Messages received: ${messageCount}`);
  console.log('='.repeat(50));
});

// Graceful shutdown
process.on('SIGINT', () => {
  console.log('\nClosing connection...');
  ws.close();
});
