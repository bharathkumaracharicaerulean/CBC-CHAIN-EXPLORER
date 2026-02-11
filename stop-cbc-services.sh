#!/bin/bash
# Stop all CBC Chain services script

echo "=========================================="
echo "Stopping CBC Chain Services"
echo "=========================================="
echo ""

# Stop explorer services
echo "Step 1: Stopping explorer services..."
cd /home/bharat/Documents/cbc-blockchain/cbc-chain-explorer
docker-compose down
cd ..
echo "✓ Explorer services stopped"
echo ""

# Stop CBC node
echo "Step 2: Stopping CBC node..."
pkill -f "cbc-node"
if [ $? -eq 0 ]; then
    echo "✓ CBC node stopped"
else
    echo "⚠ No CBC node process found (may already be stopped)"
fi
echo ""

# Clean up any database locks
echo "Step 3: Cleaning up database locks..."
rm -rf ~/.local/share/cbc-node/ 2>/dev/null
echo "✓ Database locks cleaned"
echo ""

echo "=========================================="
echo "All services stopped successfully!"
echo "=========================================="
echo ""
echo "To start services again, run:"
echo "  ./clean-restart.sh    (for fresh chain)"
echo ""
