#!/bin/bash

# Exit on error
set -e

echo "=========================================="
echo "Starting CBC Explorer Natively (No-Docker App)"
echo "=========================================="

# Step 1: Start MySQL & Redis Database Containers
echo "Step 1: Starting MySQL & Redis database containers..."
docker compose -f docker-compose.yml down || true
docker compose -f docker-compose.db.yml up -d

echo "Waiting for MySQL database container to become ready..."
until docker exec cbc-mysql mysqladmin ping -uroot -phelloload --silent; do
    echo -n "."
    sleep 1
done
sleep 3
echo ""
echo "✓ Database containers ready"

# Step 2: Compile Explorer Backend
echo "Step 2: Compiling Go explorer backend locally..."
go build -o ./bin/cbcscan -v ./cmd
echo "✓ Backend compiled successfully to ./bin/cbcscan"

# Step 3: Stop any existing local services to avoid port conflicts
echo "Step 3: Stopping any existing local explorer processes..."
pkill -f "cbcscan --conf" || true
pkill -f "next-server" || true
pkill -f "node_modules/next/dist/bin/next" || true

# Step 4: Export Environment Variables for Backend
export CONF_DIR="./configs"
export MYSQL_HOST="127.0.0.1"
export MYSQL_PASS="helloload"
export MYSQL_USER="root"
export MYSQL_DB="cbc_explorer"
export REDIS_ADDR="127.0.0.1:6379"
export CHAIN_WS_ENDPOINT="ws://127.0.0.1:9944"
export NETWORK_NODE="cbc"
export DEPLOY_ENV="dev"
export SUBSTRATE_ADDRESS_TYPE=42
export SUBSTRATE_ACCURACY=12
export DISABLE_EVM="true"
export EVM_ENABLED="false"
export METRICS_PORT="8083"
export SKIP_EVM_PLUGIN="true"
export NO_EVM="true"
export LOG_LEVEL="debug"
export DEBUG="true"

# Step 5: Launch Explorer Services on Host
echo "Step 4: Launching local observer, worker, and API..."
mkdir -p logs

# Run Observer
nohup ./bin/cbcscan --conf configs start subscribe > logs/observer.log 2>&1 &
echo "✓ Observer started (PID: $!)"

# Run Worker
nohup ./bin/cbcscan --conf configs start worker > logs/worker.log 2>&1 &
echo "✓ Worker started (PID: $!)"

# Run API Server (Default Command)
nohup ./bin/cbcscan --conf configs > logs/api.log 2>&1 &
echo "✓ API Server started (PID: $!)"

# Step 6: Launch React UI
echo "Step 5: Launching React UI frontend natively..."
cd /home/bharat/projects/CBC-EXPLORER-UI
export NEXT_PUBLIC_API_HOST="http://localhost:4399"
nohup npm run dev -- -p 3000 > react_ui.log 2>&1 &
echo "✓ React UI started on http://localhost:3000 (PID: $!)"

echo "=========================================="
echo "CBC Explorer started successfully!"
echo "=========================================="
echo "Backend Logs:"
echo "  Tail Observer: tail -f /home/bharat/projects/CBC-CHAIN-EXPLORER/logs/observer.log"
echo "  Tail API:      tail -f /home/bharat/projects/CBC-CHAIN-EXPLORER/logs/api.log"
echo "Frontend Logs:"
echo "  Tail React UI: tail -f /home/bharat/projects/CBC-EXPLORER-UI/react_ui.log"
echo "=========================================="
