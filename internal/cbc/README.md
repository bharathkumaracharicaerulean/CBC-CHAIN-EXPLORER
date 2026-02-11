# CBC Chain Integration for Subscan Explorer

This package provides CBC Chain specific initialization and runtime handling for the Subscan blockchain explorer.

## Overview

The CBC Chain integration replaces the external Python bootstrap script with native Go code that:

1. Automatically detects when the runtime_versions table is empty
2. Fetches metadata and runtime version from the CBC node via HTTP RPC
3. Populates the database with the correct runtime information
4. Verifies DCF finality integration is working
5. Integrates seamlessly with Subscan's existing architecture

## Architecture

### Components

#### `cbc_init.go`
Main initialization logic that handles:
- Runtime version detection and bootstrapping
- Metadata fetching via HTTP RPC (more reliable than WebSocket)
- Database population with retry logic
- DCF finality verification

#### `cbc_types.go`
CBC-specific type definitions and constants:
- CBC pallet names (PalletCbcPoi, PalletCbcPos, Dcf)
- Default configuration values
- Helper functions for endpoint conversion

### Integration Points

The CBC initialization is integrated into Subscan's service layer:

```go
// internal/service/service.go
func New() (s *Service) {
    // ... standard initialization ...
    
    // CBC Chain specific initialization
    if err := s.initCBCChain(); err != nil {
        util.Logger().Warning(fmt.Sprintf("CBC initialization warning: %v", err))
    }
    
    // ... continue with standard initialization ...
}
```

## How It Works

### 1. Automatic Detection

When the Subscan service starts, it checks if the network is CBC Chain:

```go
if util.NetworkNode != "cbc" && util.NetworkNode != "cbc-chain" {
    // Skip CBC initialization for other networks
    return nil
}
```

### 2. Bootstrap Check

The initializer checks if the runtime_versions table needs bootstrapping:

```go
recent := c.dao.RuntimeVersionRecent()
if recent != nil && strings.HasPrefix(recent.RawData, "0x") {
    // Already initialized, skip
    return nil
}
```

### 3. Metadata Fetching

Metadata is fetched via HTTP RPC (not WebSocket) for reliability:

```go
// HTTP is more reliable than WebSocket for large metadata payloads
req := RPCRequest{
    ID:      1,
    JSONRPC: "2.0",
    Method:  "state_getMetadata",
}
```

### 4. Database Population

Runtime version and metadata are inserted into the database:

```go
c.dao.CreateRuntimeVersion(ctx, version.SpecName, version.SpecVersion, 0)
c.dao.SetRuntimeData(version.SpecVersion, modules, metadataHex)
```

### 5. Codec Registration

The metadata is registered with Subscan's codec system:

```go
metadata.Latest(&metadata.RuntimeRaw{
    Spec: runtimeVersion.SpecVersion,
    Raw:  metadataHex,
})
```

## Configuration

The CBC initialization uses the existing Subscan configuration:

- **WebSocket Endpoint**: `util.WSEndPoint` (from config.yaml)
- **HTTP Endpoint**: Automatically derived from WebSocket endpoint
- **Network Name**: `util.NetworkNode` (should be "cbc" or "cbc-chain")
- **Database**: Uses existing DAO layer

### config.yaml Example

```yaml
server:
  http:
    addr: 0.0.0.0:4399
database:
  mysql:
    api: "mysql://root:password@127.0.0.1:3306/cbc_explorer?..."
```

The network name is typically set via environment variable or command-line flag.

## Error Handling

The implementation includes robust error handling:

### Retry Logic
- 3 retries for RPC calls
- 2-second wait between retries
- Detailed error logging

### Graceful Degradation
- CBC initialization warnings don't stop service startup
- DCF finality verification is informational only
- Falls back to standard Subscan behavior if CBC init fails

### Logging
All operations are logged with appropriate levels:
- `Info`: Normal operations
- `Warning`: Retries and non-critical issues
- `Error`: Critical failures

## Comparison with Python Bootstrap

### Python Script (Old Approach)
```python
# External script, manual execution
python3 bootstrap_cbc_runtime.py

# Pros:
- Simple, standalone
- Easy to debug

# Cons:
- Manual execution required
- Not integrated with Subscan
- Separate dependency (Python)
- No automatic retry on service restart
```

### Go Integration (New Approach)
```go
// Automatic, integrated
cbcInit := cbc.NewCBCInitializer(s.dao, util.WSEndPoint)
cbcInit.Initialize(ctx)

// Pros:
- Automatic on service start
- Native Go, no external dependencies
- Integrated with Subscan architecture
- Automatic retry logic
- Production-ready error handling

// Cons:
- More complex implementation
```

## Usage

### Automatic Usage (Recommended)

The CBC initialization runs automatically when:
1. Subscan service starts (API, Observer, Worker)
2. Network is configured as "cbc" or "cbc-chain"
3. Runtime versions table is empty or incomplete

No manual intervention required!

### Manual Verification

Check if initialization succeeded:

```bash
# Check logs
docker logs cbc-chain-explorer-observer-1 | grep "CBC"

# Expected output:
# Initializing CBC Chain specific components...
# Fetched runtime version: cbc-chain v100
# Fetched metadata: 427520 bytes
# Runtime version 100 inserted/updated successfully
# DCF finality verification passed
# CBC Chain initialization completed
```

### Database Verification

```sql
-- Check runtime_versions table
SELECT id, name, spec_version, block_num, LENGTH(raw_data) as metadata_size 
FROM runtime_versions;

-- Expected result:
-- +----+-----------+--------------+-----------+---------------+
-- | id | name      | spec_version | block_num | metadata_size |
-- +----+-----------+--------------+-----------+---------------+
-- |  1 | cbc-chain |          100 |         0 |        427520 |
-- +----+-----------+--------------+-----------+---------------+
```

## DCF Finality Verification

The initializer verifies that DCF finality is properly integrated:

```go
func (c *CBCInitializer) VerifyDCFFinality() error {
    // Checks chain_getFinalizedHead RPC
    // Warns if finalized head is genesis (indicates DCF sync issue)
    // Non-blocking - informational only
}
```

### Expected Behavior

**Healthy System:**
```
Finalized block hash: 0x1234...abcd
DCF finality verification passed
```

**DCF Sync Issue:**
```
WARNING: Finalized head appears to be genesis block
This is expected if the chain just started
If blocks are being produced but not finalized, check the finality sync task
```

## Troubleshooting

### Issue: "CBC initialization failed"

**Cause**: Cannot connect to CBC node or fetch metadata

**Solution**:
1. Check node is running: `curl http://172.17.0.1:9933`
2. Verify WebSocket endpoint in config.yaml
3. Check Docker networking (use 172.17.0.1 for host gateway)

### Issue: "Runtime version already exists, skipping bootstrap"

**Cause**: Database already has runtime version

**Solution**: This is normal! The bootstrap only runs once.

### Issue: "DCF finality verification warning"

**Cause**: Finalized head is genesis block

**Solution**:
1. If chain just started: This is expected, wait for blocks
2. If blocks are producing: Check finality sync task in cbc-node/src/service.rs
3. This is informational only - doesn't block indexing

### Issue: Build errors

**Cause**: Missing dependencies or import issues

**Solution**:
```bash
cd cbc-chain-explorer
go mod tidy
go build ./cmd
```

## Testing

### Unit Testing

```bash
cd cbc-chain-explorer/internal/cbc
go test -v
```

### Integration Testing

```bash
# Start CBC node
./start-cbc-node.sh

# Start explorer (will auto-initialize)
cd cbc-chain-explorer
docker-compose up -d

# Check logs
docker logs -f cbc-chain-explorer-observer-1

# Verify API
curl -X POST http://localhost:4399/api/scan/metadata
curl -X POST http://localhost:4399/api/scan/blocks
```

## Future Enhancements

Potential improvements for future versions:

1. **Custom Type Registration**: Automatically register CBC-specific types from cbc.json
2. **Health Checks**: Periodic DCF finality monitoring
3. **Metrics**: Prometheus metrics for CBC-specific operations
4. **Configuration**: CBC-specific config section in config.yaml
5. **Migration Tool**: Automatic migration from Python bootstrap to Go

## Related Files

- `cbc-chain-explorer/internal/service/service.go` - Service initialization
- `cbc-chain-explorer/internal/dao/runtimeVersion.go` - Database operations
- `cbc-chain-explorer/configs/source/cbc.json` - CBC type definitions
- `cbc-chain/cbc-node/src/service.rs` - DCF finality sync task
- `bootstrap_cbc_runtime.py` - Legacy Python bootstrap (deprecated)

## License

Same as Subscan Explorer (Apache 2.0)
