# CBC Chain Explorer - Usage Guide

Complete guide for running and managing the CBC blockchain explorer.

---

## Table of Contents

1. [Quick Start](#quick-start)
2. [Manual Service Management](#manual-service-management)
3. [Script-Based Management](#script-based-management)
4. [Monitoring & Troubleshooting](#monitoring--troubleshooting)
5. [API Usage](#api-usage)

---

## Quick Start

### Prerequisites

- CBC node binary built: `../cbc-chain/target/release/cbc-node`
- MySQL running with database `cbc_explorer`
- Redis running on port 6379
- Docker and docker-compose installed

### Fastest Way to Start

```bash
cd /home/bharat/Documents/cbc-blockchain/cbc-chain-explorer
./clean-restart.sh
```

Then open browser to http://localhost:3000 and do a **hard refresh** (Ctrl+Shift+R).

---

## Manual Service Management

### 1. Manual Startup (Step-by-Step)

#### Step 1: Prepare Database

```bash
# Create/reset database
mysql -uroot -phelloload -e "DROP DATABASE IF EXISTS cbc_explorer; CREATE DATABASE cbc_explorer;"

# Clear Redis cache
redis-cli FLUSHALL
```

#### Step 2: Start CBC Node

```bash
# Navigate to CBC chain directory
cd /home/bharat/Documents/cbc-blockchain/cbc-chain

# Start the node
nohup ./target/release/cbc-node \
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
  > cbc-node.log 2>&1 &

# Note the process ID
echo $!
```

#### Step 3: Verify Node is Ready

```bash
# Wait for node to be ready (check every 2 seconds)
while ! curl -s -H "Content-Type: application/json" \
  -d '{"id":1, "jsonrpc":"2.0", "method": "system_health"}' \
  http://localhost:9933 > /dev/null 2>&1; do
  echo "Waiting for node..."
  sleep 2
done
echo "Node is ready!"
```

#### Step 4: Start Explorer Services

```bash
# Navigate to explorer directory
cd /home/bharat/Documents/cbc-blockchain/cbc-chain-explorer

# Start all services
docker-compose up -d

# Check status
docker-compose ps
```

#### Step 5: Verify Explorer is Running

```bash
# Wait 30 seconds for initialization
sleep 30

# Check observer logs
docker logs cbc-chain-explorer_cbc-explorer-observer_1 2>&1 | tail -20

# Check blocks are being indexed
mysql -uroot -phelloload cbc_explorer -e "SELECT COUNT(*) FROM chain_blocks;"

# Check API
curl -s -X POST http://localhost:4399/api/scan/metadata | jq '.'
```

### 2. Manual Shutdown (Step-by-Step)

#### Step 1: Stop Explorer Services

```bash
cd /home/bharat/Documents/cbc-blockchain/cbc-chain-explorer
docker-compose down
```

#### Step 2: Stop CBC Node

```bash
# Find and kill the node process
pkill -f "cbc-node"

# Verify it's stopped
ps aux | grep cbc-node | grep -v grep
```

#### Step 3: Clean Up (Optional)

```bash
# Remove database locks
rm -rf ~/.local/share/cbc-node/

# Clear Redis (if starting fresh next time)
redis-cli FLUSHALL
```

---

## Script-Based Management

### Available Scripts

#### 1. `clean-restart.sh` - Fresh Chain Restart

**Purpose:** Complete fresh restart with clean database and cache

**Usage:**
```bash
cd /home/bharat/Documents/cbc-blockchain/cbc-chain-explorer
./clean-restart.sh
```

**What it does:**
1. ✓ Stops all services (explorer + node)
2. ✓ Resets database (drops and recreates)
3. ✓ Clears Redis cache
4. ✓ Starts CBC node with readiness check (waits up to 60 seconds)
5. ✓ Starts explorer services
6. ✓ Verifies block indexing

**Expected output:**
```
==========================================
CBC Chain - Clean Restart
==========================================

Step 1: Stopping all services...
✓ Services stopped

Step 2: Resetting database...
✓ Database reset

Step 3: Clearing Redis cache...
✓ Redis cache cleared

Step 4: Starting CBC node...
✓ CBC node started (PID: XXXXX)

Step 5: Waiting for CBC node to be ready...
.....✓ CBC node is ready (took X seconds)

Step 6: Starting explorer services...
✓ Explorer started

Step 7: Waiting for observer initialization (30 seconds)...

Step 8: Checking observer status...
✓ Observer started successfully

Step 9: Waiting for block indexing (20 seconds)...

Step 10: Verifying block indexing...
Blocks indexed: X
✓ Blocks are being indexed successfully

==========================================
Clean restart complete!
==========================================
```

**After running:**
- Open browser to http://localhost:3000
- Do a **hard refresh** (Ctrl+Shift+R) to clear browser cache

---

#### 2. `stop-cbc-services.sh` - Stop All Services

**Purpose:** Cleanly stop all running services

**Usage:**
```bash
cd /home/bharat/Documents/cbc-blockchain/cbc-chain-explorer
./stop-cbc-services.sh
```

**What it does:**
1. ✓ Stops explorer services (docker-compose down)
2. ✓ Stops CBC node process
3. ✓ Cleans up database locks

**Expected output:**
```
==========================================
Stopping CBC Chain Services
==========================================

Step 1: Stopping explorer services...
✓ Explorer services stopped

Step 2: Stopping CBC node...
✓ CBC node stopped

Step 3: Cleaning up database locks...
✓ Database locks cleaned

==========================================
All services stopped successfully!
==========================================
```

---

## Monitoring & Troubleshooting

### Check Service Status

```bash
# Check explorer services
cd /home/bharat/Documents/cbc-blockchain/cbc-chain-explorer
docker-compose ps

# Check CBC node
ps aux | grep cbc-node | grep -v grep

# Check all together
docker-compose ps && echo "---" && ps aux | grep cbc-node | grep -v grep
```

### View Logs

```bash
# Observer logs (real-time)
docker logs -f cbc-chain-explorer_cbc-explorer-observer_1

# Observer logs (last 50 lines)
docker logs --tail 50 cbc-chain-explorer_cbc-explorer-observer_1

# Worker logs
docker logs -f cbc-chain-explorer_cbc-explorer-worker_1

# API logs
docker logs -f cbc-chain-explorer_cbc-explorer-api_1

# CBC node logs
tail -f /home/bharat/Documents/cbc-blockchain/cbc-chain/cbc-node.log

# All explorer services
docker-compose logs -f
```

### Check Database

```bash
# Block count
mysql -uroot -phelloload cbc_explorer -e "SELECT COUNT(*) as total_blocks FROM chain_blocks;"

# Latest blocks
mysql -uroot -phelloload cbc_explorer -e "SELECT block_num, block_timestamp, extrinsics_count, event_count FROM chain_blocks ORDER BY block_num DESC LIMIT 10;"

# Extrinsic count
mysql -uroot -phelloload cbc_explorer -e "SELECT COUNT(*) as total_extrinsics FROM chain_extrinsics;"

# Runtime versions
mysql -uroot -phelloload cbc_explorer -e "SELECT spec_version, block_num, LENGTH(raw_data) as metadata_size FROM runtime_versions;"
```

### Check Redis Cache

```bash
# View all keys
redis-cli KEYS "*"

# View metadata
redis-cli HGETALL "cbc:metadata"

# Check specific values
redis-cli HGET "cbc:metadata" "count_extrinsic"
redis-cli HGET "cbc:metadata" "finalized_blockNum"
```

### Common Issues & Solutions

#### Issue 1: Observer keeps restarting

**Symptoms:**
```bash
docker logs cbc-chain-explorer_cbc-explorer-observer_1 | grep panic
# Shows: "panic: Can not find chain metadata"
```

**Solution:**
```bash
# Node wasn't ready when observer started
./stop-cbc-services.sh
./clean-restart.sh  # This waits for node readiness
```

#### Issue 2: No blocks being indexed

**Check:**
```bash
# Is node producing blocks?
curl -s -H "Content-Type: application/json" \
  -d '{"id":1, "jsonrpc":"2.0", "method": "chain_getBlock"}' \
  http://localhost:9933 | jq '.result.block.header.number'

# Is observer subscribed?
docker logs cbc-chain-explorer_cbc-explorer-observer_1 | grep subscribe
# Should show: "subscribe from chain success!"
```

**Solution:**
```bash
# Restart observer
docker-compose restart cbc-explorer-observer
```

#### Issue 3: Browser shows old data (stale cache)

**Symptoms:** UI shows old extrinsic counts or block numbers

**Solution:**
- Do a **hard refresh** in browser: Ctrl+Shift+R (Linux/Windows) or Cmd+Shift+R (Mac)
- Or clear browser cache for localhost:3000

#### Issue 4: API returns empty data

**Check:**
```bash
# Check API is running
curl -s http://localhost:4399/api/scan/metadata | jq '.'

# Check database has data
mysql -uroot -phelloload cbc_explorer -e "SELECT COUNT(*) FROM chain_blocks;"
```

**Solution:**
```bash
# Wait for indexing (may take 1-2 minutes after startup)
sleep 60
# Then check again
```

---

## API Usage

### Base URL
```
http://localhost:4399/api/scan
```

### Common Endpoints

#### 1. Get Metadata
```bash
curl -X POST http://localhost:4399/api/scan/metadata -d '{}'
```

**Response:**
```json
{
  "code": 0,
  "message": "Success",
  "data": {
    "count_extrinsic": "22",
    "count_signed_extrinsic": "0",
    "finalized_blockNum": "25",
    "networkNode": "cbc",
    "total_account": "0",
    "total_transfer": "0"
  }
}
```

#### 2. Get Blocks
```bash
curl -X POST http://localhost:4399/api/scan/blocks -d '{"row":10,"page":0}'
```

**Response:**
```json
{
  "code": 0,
  "message": "Success",
  "data": {
    "blocks": [
      {
        "block_num": 25,
        "block_timestamp": 1770801234,
        "hash": "0x...",
        "extrinsics_count": 1,
        "event_count": 7
      }
    ],
    "count": 25
  }
}
```

#### 3. Get Block by Number
```bash
curl -X POST http://localhost:4399/api/scan/block -d '{"block_num":10}'
```

#### 4. Get Extrinsics
```bash
curl -X POST http://localhost:4399/api/scan/extrinsics -d '{"row":10,"page":0}'
```

#### 5. Get Extrinsic by Index
```bash
curl -X POST http://localhost:4399/api/scan/extrinsic -d '{"extrinsic_index":"10-0"}'
```

### Using with jq (Pretty Print)

```bash
# Pretty print metadata
curl -s -X POST http://localhost:4399/api/scan/metadata -d '{}' | jq '.'

# Get specific field
curl -s -X POST http://localhost:4399/api/scan/metadata -d '{}' | jq '.data.count_extrinsic'

# Get latest 5 blocks
curl -s -X POST http://localhost:4399/api/scan/blocks -d '{"row":5,"page":0}' | jq '.data.blocks[]'
```

---

## Configuration

### Environment Variables

Edit `docker-compose.yml` to modify:

```yaml
environment:
  MYSQL_HOST: 127.0.0.1
  MYSQL_PASS: 'helloload'
  MYSQL_USER: 'root'
  MYSQL_DB: 'cbc_explorer'
  REDIS_ADDR: 127.0.0.1:6379
  CHAIN_WS_ENDPOINT: 'ws://172.17.0.1:9933'
  NETWORK_NODE: 'cbc'
  DEPLOY_ENV: 'dev'
  LOG_LEVEL: 'debug'
```

### Custom Types

CBC chain custom types are defined in:
```
configs/source/cbc.json
```

---

## Quick Reference Commands

### Start Services
```bash
# Using script (recommended)
./clean-restart.sh

# Manual
# 1. Start node: cd ../cbc-chain && nohup ./target/release/cbc-node --dev ... &
# 2. Start explorer: docker-compose up -d
```

### Stop Services
```bash
# Using script
./stop-cbc-services.sh

# Manual
docker-compose down && pkill -f cbc-node
```

### Check Status
```bash
docker-compose ps
ps aux | grep cbc-node
mysql -uroot -phelloload cbc_explorer -e "SELECT COUNT(*) FROM chain_blocks;"
```

### View Logs
```bash
docker logs -f cbc-chain-explorer_cbc-explorer-observer_1
tail -f ../cbc-chain/cbc-node.log
```

### Access UI
```
http://localhost:3000
```

### Access API
```
http://localhost:4399/api/scan/metadata
```

---

## Support

For issues or questions:
1. Check logs: `docker-compose logs -f`
2. Verify node is running: `curl http://localhost:9933`
3. Check database: `mysql -uroot -phelloload cbc_explorer`
4. Review this guide's troubleshooting section
