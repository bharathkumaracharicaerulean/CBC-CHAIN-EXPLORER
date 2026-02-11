#!/bin/bash
# Clean restart script for CBC Chain system

echo "=========================================="
echo "CBC Chain - Clean Restart"
echo "=========================================="
echo ""

# Stop all services
echo "Step 1: Stopping all services..."
docker-compose down
cd ..
pkill -f "cbc-node" || true
sleep 2
echo "✓ Services stopped"
echo ""

# Drop and recreate database
echo "Step 2: Resetting database..."
mysql -uroot -phelloload -e "DROP DATABASE IF EXISTS cbc_explorer; CREATE DATABASE cbc_explorer;" 2>&1 | grep -v Warning
echo "✓ Database reset"
echo ""

# Clear Redis cache
echo "Step 3: Clearing Redis cache..."
redis-cli FLUSHALL > /dev/null 2>&1
if [ $? -eq 0 ]; then
    echo "✓ Redis cache cleared"
else
    echo "⚠ Warning: Could not clear Redis cache (redis-cli may not be available)"
fi
echo ""

# Start node
echo "Step 4: Starting CBC node..."
# Get paths
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
CBC_DIR="$SCRIPT_DIR/cbc-chain"
CBC_BINARY="$CBC_DIR/target/release/cbc-node"
LOG_FILE="$CBC_DIR/cbc-node.log"

# Check if the binary exists
if [ ! -f "$CBC_BINARY" ]; then
    echo "❌ Error: CBC node binary not found at: $CBC_BINARY"
    echo "   Please build the node first: cd cbc-chain && cargo build --release"
    exit 1
fi

# Start the node
cd "$CBC_DIR"
nohup "$CBC_BINARY" \
  --dev \
  --validator \
  --alice \
  --tmp \
  --rpc-port 9933 \
  --rpc-cors all \
  --rpc-external \
  --unsafe-rpc-external \
  --enable-cbc-extensions \
  --cbc-mode development \
  --force-authoring \
  --rpc-methods unsafe \
  --rpc-max-connections 5000 \
  --rpc-rate-limit-requests 10000 \
  --rpc-max-request-size 20 \
  --rpc-max-response-size 20 \
  --quiet \
  > /dev/null 2>&1 &

NODE_PID=$!
cd "$SCRIPT_DIR"
echo "✓ CBC node started (PID: $NODE_PID)"
echo ""

# Wait for node to be ready
echo "Step 5: Waiting for CBC node to be ready..."
MAX_RETRIES=30
RETRY_COUNT=0
NODE_READY=false

while [ $RETRY_COUNT -lt $MAX_RETRIES ]; do
    # Check if node RPC is responding
    if curl -s -H "Content-Type: application/json" -d '{"id":1, "jsonrpc":"2.0", "method": "system_health"}' http://localhost:9933 > /dev/null 2>&1; then
        NODE_READY=true
        echo "✓ CBC node is ready (took $((RETRY_COUNT * 2)) seconds)"
        break
    fi
    RETRY_COUNT=$((RETRY_COUNT + 1))
    echo -n "."
    sleep 2
done

if [ "$NODE_READY" = false ]; then
    echo ""
    echo "❌ Error: CBC node failed to become ready after $((MAX_RETRIES * 2)) seconds"
    echo "   Check node logs with: journalctl -f or check running processes"
    exit 1
fi
echo ""

# Start explorer
echo "Step 6: Starting explorer services..."
cd cbc-chain-explorer
docker-compose up -d
cd ..
echo "✓ Explorer started"
echo ""

# Wait for observer initialization
echo "Step 7: Waiting for observer initialization (30 seconds)..."
sleep 30
echo ""

# Check observer status
echo "Step 8: Checking observer status..."
echo "----------------------------------------"
if docker logs cbc-chain-explorer_cbc-explorer-observer_1 2>&1 | grep -q "panic"; then
    echo "⚠ Warning: Observer encountered errors during startup"
    docker logs cbc-chain-explorer_cbc-explorer-observer_1 2>&1 | grep -i "error\|panic" | tail -5
else
    echo "✓ Observer started successfully"
    docker logs cbc-chain-explorer_cbc-explorer-observer_1 2>&1 | grep "CBC\|subscribe" | tail -5
fi
echo "----------------------------------------"
echo ""

# Wait a bit more for blocks to be indexed
echo "Step 9: Waiting for block indexing (20 seconds)..."
sleep 20
echo ""

# Check blocks
echo "Step 10: Verifying block indexing..."
BLOCK_COUNT=$(mysql -uroot -phelloload cbc_explorer -se "SELECT COUNT(*) FROM chain_blocks;" 2>/dev/null || echo "0")
echo "Blocks indexed: $BLOCK_COUNT"

if [ "$BLOCK_COUNT" -gt 0 ]; then
    echo "✓ Blocks are being indexed successfully"
else
    echo "⚠ Warning: No blocks indexed yet. This may be normal if the chain just started."
    echo "   Wait a minute and check: mysql -uroot -phelloload cbc_explorer -e 'SELECT COUNT(*) FROM chain_blocks;'"
fi
echo ""

echo "=========================================="
echo "Clean restart complete!"
echo "=========================================="
echo ""
echo "Next steps:"
echo "  - Open browser: http://localhost:3000"
echo "  - Do a HARD REFRESH (Ctrl+Shift+R) to clear browser cache"
echo "  - View logs: docker logs -f cbc-chain-explorer_cbc-explorer-observer_1"
echo "  - Check API: curl -X POST http://localhost:4399/api/scan/blocks -d '{}' | jq '.'"
echo "  - Check blocks: mysql -uroot -phelloload cbc_explorer -e 'SELECT COUNT(*) FROM chain_blocks;'"
echo ""
