package cbc

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/itering/subscan/internal/dao"
	"github.com/itering/subscan/util"
	"github.com/itering/substrate-api-rpc/metadata"
)

// CBCInitializer handles CBC Chain specific initialization
type CBCInitializer struct {
	dao       dao.IDao
	nodeURL   string
	httpURL   string
	retries   int
	retryWait time.Duration
}

// NewCBCInitializer creates a new CBC initializer
func NewCBCInitializer(d dao.IDao, wsEndpoint string) *CBCInitializer {
	// Convert WebSocket endpoint to HTTP for more reliable metadata fetching
	httpURL := strings.Replace(wsEndpoint, "ws://", "http://", 1)
	httpURL = strings.Replace(httpURL, "wss://", "https://", 1)
	
	return &CBCInitializer{
		dao:       d,
		nodeURL:   wsEndpoint,
		httpURL:   httpURL,
		retries:   3,
		retryWait: 2 * time.Second,
	}
}

// RPCRequest represents a JSON-RPC request
type RPCRequest struct {
	ID      int           `json:"id"`
	JSONRPC string        `json:"jsonrpc"`
	Method  string        `json:"method"`
	Params  []interface{} `json:"params,omitempty"`
}

// RPCResponse represents a JSON-RPC response
type RPCResponse struct {
	ID      int             `json:"id"`
	JSONRPC string          `json:"jsonrpc"`
	Result  json.RawMessage `json:"result,omitempty"`
	Error   *RPCError       `json:"error,omitempty"`
}

// RPCError represents a JSON-RPC error
type RPCError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

// RuntimeVersion represents the runtime version info
type RuntimeVersion struct {
	SpecName         string   `json:"specName"`
	ImplName         string   `json:"implName"`
	AuthoringVersion int      `json:"authoringVersion"`
	SpecVersion      int      `json:"specVersion"`
	ImplVersion      int      `json:"implVersion"`
	Apis             [][]interface{} `json:"apis"`
	TransactionVersion int    `json:"transactionVersion"`
	StateVersion     int      `json:"stateVersion"`
}

// Initialize performs CBC-specific initialization
func (c *CBCInitializer) Initialize(ctx context.Context) error {
	util.Logger().Info("Starting CBC Chain initialization...")
	
	// Check if runtime_versions table is empty
	recent := c.dao.RuntimeVersionRecent()
	if recent != nil && strings.HasPrefix(recent.RawData, "0x") {
		util.Logger().Info(fmt.Sprintf("Runtime version already exists (spec: %d), skipping bootstrap", recent.SpecVersion))
		return nil
	}
	
	util.Logger().Info("Runtime versions table is empty or incomplete, bootstrapping...")
	
	// Fetch runtime version info
	runtimeVersion, err := c.fetchRuntimeVersion()
	if err != nil {
		return fmt.Errorf("failed to fetch runtime version: %w", err)
	}
	
	util.Logger().Info(fmt.Sprintf("Fetched runtime version: %s v%d", runtimeVersion.SpecName, runtimeVersion.SpecVersion))
	
	// Fetch metadata
	metadataHex, err := c.fetchMetadata()
	if err != nil {
		return fmt.Errorf("failed to fetch metadata: %w", err)
	}
	
	util.Logger().Info(fmt.Sprintf("Fetched metadata: %d bytes", len(metadataHex)))
	
	// Insert into database
	if err := c.insertRuntimeVersion(ctx, runtimeVersion, metadataHex); err != nil {
		return fmt.Errorf("failed to insert runtime version: %w", err)
	}
	
	// Register metadata with the codec
	metadata.Latest(&metadata.RuntimeRaw{
		Spec: runtimeVersion.SpecVersion,
		Raw:  metadataHex,
	})
	
	util.Logger().Info("CBC Chain initialization completed successfully")
	return nil
}

// fetchMetadata fetches metadata from CBC node via HTTP RPC
func (c *CBCInitializer) fetchMetadata() (string, error) {
	var lastErr error
	
	for i := 0; i < c.retries; i++ {
		if i > 0 {
			util.Logger().Warning(fmt.Sprintf("Retry %d/%d: Fetching metadata...", i, c.retries))
			time.Sleep(c.retryWait)
		}
		
		req := RPCRequest{
			ID:      1,
			JSONRPC: "2.0",
			Method:  "state_getMetadata",
		}
		
		resp, err := c.makeRPCCall(req)
		if err != nil {
			lastErr = err
			continue
		}
		
		var metadataHex string
		if err := json.Unmarshal(resp.Result, &metadataHex); err != nil {
			lastErr = fmt.Errorf("failed to unmarshal metadata: %w", err)
			continue
		}
		
		if !strings.HasPrefix(metadataHex, "0x") {
			lastErr = fmt.Errorf("invalid metadata format: missing 0x prefix")
			continue
		}
		
		return metadataHex, nil
	}
	
	return "", fmt.Errorf("failed after %d retries: %w", c.retries, lastErr)
}

// fetchRuntimeVersion fetches runtime version info from CBC node
func (c *CBCInitializer) fetchRuntimeVersion() (*RuntimeVersion, error) {
	var lastErr error
	
	for i := 0; i < c.retries; i++ {
		if i > 0 {
			util.Logger().Warning(fmt.Sprintf("Retry %d/%d: Fetching runtime version...", i, c.retries))
			time.Sleep(c.retryWait)
		}
		
		req := RPCRequest{
			ID:      1,
			JSONRPC: "2.0",
			Method:  "state_getRuntimeVersion",
		}
		
		resp, err := c.makeRPCCall(req)
		if err != nil {
			lastErr = err
			continue
		}
		
		var version RuntimeVersion
		if err := json.Unmarshal(resp.Result, &version); err != nil {
			lastErr = fmt.Errorf("failed to unmarshal runtime version: %w", err)
			continue
		}
		
		return &version, nil
	}
	
	return nil, fmt.Errorf("failed after %d retries: %w", c.retries, lastErr)
}

// makeRPCCall makes an HTTP JSON-RPC call to the CBC node
func (c *CBCInitializer) makeRPCCall(req RPCRequest) (*RPCResponse, error) {
	reqBody, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}
	
	httpReq, err := http.NewRequest("POST", c.httpURL, strings.NewReader(string(reqBody)))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	
	httpReq.Header.Set("Content-Type", "application/json")
	
	client := &http.Client{Timeout: 30 * time.Second}
	httpResp, err := client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer httpResp.Body.Close()
	
	if httpResp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", httpResp.StatusCode)
	}
	
	body, err := io.ReadAll(httpResp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}
	
	var resp RPCResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}
	
	if resp.Error != nil {
		return nil, fmt.Errorf("RPC error: %s (code: %d)", resp.Error.Message, resp.Error.Code)
	}
	
	return &resp, nil
}

// insertRuntimeVersion inserts runtime version into database
func (c *CBCInitializer) insertRuntimeVersion(ctx context.Context, version *RuntimeVersion, metadataHex string) error {
	util.Logger().Info(fmt.Sprintf("Inserting runtime version %d into database...", version.SpecVersion))
	
	// CBC Chain modules (based on our custom pallets)
	modules := "System|Timestamp|Balances|TransactionPayment|Sudo|PalletCbcPoi|PalletCbcPos|Dcf"
	
	// First, check if the data already exists and is correct
	recent := c.dao.RuntimeVersionRecent()
	if recent != nil && recent.SpecVersion == version.SpecVersion && len(recent.RawData) > 100 {
		util.Logger().Info(fmt.Sprintf("Runtime version %d already exists with %d bytes of metadata - skipping insert", version.SpecVersion, len(recent.RawData)))
		return nil
	}
	
	// Try to create the runtime version entry
	created := c.dao.CreateRuntimeVersion(ctx, version.SpecName, version.SpecVersion, 0)
	util.Logger().Info(fmt.Sprintf("CreateRuntimeVersion returned: %v", created))
	
	// Set the runtime data
	affected := c.dao.SetRuntimeData(version.SpecVersion, modules, metadataHex)
	util.Logger().Info(fmt.Sprintf("SetRuntimeData affected %d rows", affected))
	
	// Verify the data was inserted successfully
	recent = c.dao.RuntimeVersionRecent()
	if recent == nil || recent.SpecVersion != version.SpecVersion || len(recent.RawData) < 100 {
		return fmt.Errorf("failed to verify runtime data after insert - metadata not found in database")
	}
	
	util.Logger().Info(fmt.Sprintf("Runtime version %d inserted/updated successfully with %d bytes of metadata", version.SpecVersion, len(recent.RawData)))
	return nil
}

// VerifyDCFFinality checks if DCF finality is working properly
func (c *CBCInitializer) VerifyDCFFinality() error {
	util.Logger().Info("Verifying DCF finality integration...")
	
	req := RPCRequest{
		ID:      1,
		JSONRPC: "2.0",
		Method:  "chain_getFinalizedHead",
	}
	
	resp, err := c.makeRPCCall(req)
	if err != nil {
		return fmt.Errorf("failed to get finalized head: %w", err)
	}
	
	var finalizedHash string
	if err := json.Unmarshal(resp.Result, &finalizedHash); err != nil {
		return fmt.Errorf("failed to unmarshal finalized hash: %w", err)
	}
	
	// Get block number for finalized hash
	req = RPCRequest{
		ID:      2,
		JSONRPC: "2.0",
		Method:  "chain_getBlock",
		Params:  []interface{}{finalizedHash},
	}
	
	resp, err = c.makeRPCCall(req)
	if err != nil {
		return fmt.Errorf("failed to get finalized block: %w", err)
	}
	
	// Parse block to check if it's not genesis
	var blockData map[string]interface{}
	if err := json.Unmarshal(resp.Result, &blockData); err != nil {
		return fmt.Errorf("failed to unmarshal block data: %w", err)
	}
	
	util.Logger().Info(fmt.Sprintf("Finalized block hash: %s", finalizedHash))
	
	// Check if finalized block is genesis (indicates DCF finality issue)
	if finalizedHash == "0x" || strings.HasSuffix(finalizedHash, "000000000000000000000000000000000000000000000000000000000000") {
		util.Logger().Warning("WARNING: Finalized head appears to be genesis block - DCF finality may not be syncing properly")
		util.Logger().Warning("This is expected if the chain just started. If blocks are being produced but not finalized, check the finality sync task in cbc-node/src/service.rs")
	} else {
		util.Logger().Info("DCF finality verification passed")
	}
	
	return nil
}
