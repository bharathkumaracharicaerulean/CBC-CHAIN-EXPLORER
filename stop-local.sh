#!/bin/bash

echo "=========================================="
echo "Stopping Local CBC Explorer Services"
echo "=========================================="

# Step 1: Stop Native Go Explorer processes
echo "Step 1: Stopping local Go explorer processes..."
pkill -f "cbcscan --conf" || true
echo "✓ Explorer processes stopped"

# Step 2: Stop React UI processes
echo "Step 2: Stopping React UI frontend..."
pkill -f "next-server" || true
pkill -f "node_modules/next/dist/bin/next" || true
echo "✓ React UI frontend stopped"

# Step 3: Stop MySQL and Redis containers
echo "Step 3: Stopping database containers..."
docker compose -f /home/bharat/projects/CBC-CHAIN-EXPLORER/docker-compose.db.yml down || true
echo "✓ Database containers stopped"

echo "=========================================="
echo "All local explorer services stopped successfully!"
echo "=========================================="
