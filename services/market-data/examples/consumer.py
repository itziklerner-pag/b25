#!/usr/bin/env python3
"""
Example consumer for Market Data Service
Subscribes to order book updates via Redis pub/sub
"""

import redis
import json
import sys
from datetime import datetime

def format_orderbook(data):
    """Format order book for display"""
    symbol = data.get('symbol', 'UNKNOWN')
    timestamp = data.get('timestamp', 0)
    dt = datetime.fromtimestamp(timestamp / 1_000_000)  # microseconds to seconds

    print(f"\n{'='*60}")
    print(f"Symbol: {symbol} | Time: {dt.strftime('%H:%M:%S.%f')}")
    print(f"{'='*60}")

    bids = data.get('bids', {})
    asks = data.get('asks', {})

    # Get top 5 levels
    print("\nAsks (sellers):")
    ask_items = sorted(asks.items(), key=lambda x: float(x[0]), reverse=True)[:5]
    for price, qty in reversed(ask_items):
        print(f"  {float(price):>12,.2f} | {float(qty):>10,.4f}")

    print("\n" + "-"*40)

    print("\nBids (buyers):")
    bid_items = sorted(bids.items(), key=lambda x: float(x[0]), reverse=True)[:5]
    for price, qty in bid_items:
        print(f"  {float(price):>12,.2f} | {float(qty):>10,.4f}")

    # Calculate spread
    if bid_items and ask_items:
        best_bid = float(bid_items[0][0])
        best_ask = float(sorted(asks.items(), key=lambda x: float(x[0]))[0][0])
        spread = best_ask - best_bid
        spread_bps = (spread / best_bid) * 10000
        print(f"\nSpread: ${spread:.2f} ({spread_bps:.2f} bps)")

def format_trade(data):
    """Format trade for display"""
    symbol = data.get('symbol', 'UNKNOWN')
    price = data.get('price', 0)
    quantity = data.get('quantity', 0)
    timestamp = data.get('timestamp', 0)
    is_buyer_maker = data.get('is_buyer_maker', False)
    side = 'SELL' if is_buyer_maker else 'BUY'

    dt = datetime.fromtimestamp(timestamp / 1000)  # milliseconds to seconds

    print(f"[{dt.strftime('%H:%M:%S')}] {symbol} {side:4s} {quantity:>10,.4f} @ ${price:>10,.2f}")

def main():
    symbol = sys.argv[1] if len(sys.argv) > 1 else 'BTCUSDT'

    print(f"Connecting to Redis...")
    r = redis.Redis(host='localhost', port=6379, decode_responses=True)

    # Test connection
    try:
        r.ping()
        print(f"Connected to Redis successfully")
    except redis.ConnectionError:
        print("ERROR: Could not connect to Redis. Is it running?")
        print("Start it with: docker-compose up -d redis")
        sys.exit(1)

    pubsub = r.pubsub()

    # Subscribe to both order book and trades
    orderbook_channel = f'orderbook:{symbol}'
    trade_channel = f'trades:{symbol}'

    print(f"\nSubscribing to:")
    print(f"  - {orderbook_channel}")
    print(f"  - {trade_channel}")
    print(f"\nWaiting for messages... (Ctrl+C to quit)\n")

    pubsub.subscribe(orderbook_channel, trade_channel)

    try:
        for message in pubsub.listen():
            if message['type'] != 'message':
                continue

            channel = message['channel']
            data = json.loads(message['data'])

            if 'orderbook' in channel:
                format_orderbook(data)
            elif 'trades' in channel:
                format_trade(data)

    except KeyboardInterrupt:
        print("\n\nShutting down...")
        pubsub.unsubscribe()
        r.close()

if __name__ == '__main__':
    main()
