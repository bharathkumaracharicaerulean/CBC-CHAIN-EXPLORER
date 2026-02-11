# CBC Explorer - Quick Reference

##  Quick Start

```bash
cd /home/bharat/Documents/cbc-blockchain/cbc-chain-explorer
./clean-restart.sh
```
Then open http://localhost:3000 and press **Ctrl+Shift+R**

---

##  Scripts

### Start Fresh Chain
```bash
./clean-restart.sh
```
Stops everything → Resets DB → Clears cache → Starts node → Starts explorer

### Stop All Services
```bash
./stop-cbc-services.sh
```
Stops explorer → Stops node → Cleans locks

---

##  Manual Commands

### Start Services
```bash
# 1. Reset database
mysql -uroot -phelloload -e "DROP DATABASE IF EXISTS cbc_explorer; CREATE DATABASE cbc_explorer;"
redis-cli FLUSHALL

# 2. Start node
cd ../cbc-chain
nohup ./target/release/cbc-node --dev --validator --alice --tmp \
  --rpc-port 9933 --rpc-cors all --rpc-external --unsafe-rpc-external \
  --enable-cbc-extensions --cbc-mode development --force-authoring \
  --rpc-methods unsafe > cbc-node.log 2>&1 &

# 3. Wait for node
while ! curl -s http://localhost:9933 > /dev/null; do sleep 2; done

# 4. Start explorer
cd ../cbc-chain-explorer
docker-compose up -d
```

### Stop Services
```bash
docker-compose down
pkill -f cbc-node
```

---

##  Monitoring

### Check Status
```bash
docker-compose ps                    # Explorer services
ps aux | grep cbc-node              # Node process
```

### View Logs
```bash
docker logs -f cbc-chain-explorer_cbc-explorer-observer_1  # Observer
tail -f ../cbc-chain/cbc-node.log                          # Node
```

### Check Data
```bash
# Blocks indexed
mysql -uroot -phelloload cbc_explorer -e "SELECT COUNT(*) FROM chain_blocks;"

# Latest blocks
mysql -uroot -phelloload cbc_explorer -e "SELECT block_num, extrinsics_count FROM chain_blocks ORDER BY block_num DESC LIMIT 5;"

# Redis metadata
redis-cli HGETALL "cbc:metadata"
```

---

##  Access Points

| Service | URL |
|---------|-----|
| **Web UI** | http://localhost:3000 |
| **API** | http://localhost:4399/api/scan |
| **Node RPC** | http://localhost:9933 |

---

##  API Examples

```bash
# Metadata
curl -X POST http://localhost:4399/api/scan/metadata -d '{}'

# Blocks
curl -X POST http://localhost:4399/api/scan/blocks -d '{"row":10,"page":0}'

# Specific block
curl -X POST http://localhost:4399/api/scan/block -d '{"block_num":10}'

# Extrinsics
curl -X POST http://localhost:4399/api/scan/extrinsics -d '{"row":10,"page":0}'
```

---

##  Common Issues

| Problem | Solution |
|---------|----------|
| **Browser shows old data** | Hard refresh: **Ctrl+Shift+R** |
| **Observer keeps restarting** | Run `./clean-restart.sh` (waits for node) |
| **No blocks indexed** | Check: `docker logs cbc-chain-explorer_cbc-explorer-observer_1` |
| **API returns empty** | Wait 1-2 minutes for indexing to start |

---

## 📖 Full Documentation

See [USAGE.md](file:///home/bharat/Documents/cbc-blockchain/cbc-chain-explorer/USAGE.md) for detailed instructions.
