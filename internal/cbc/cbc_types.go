package cbc

// CBC Chain specific type definitions and constants

const (
	// CBCChainName is the name of the CBC blockchain
	CBCChainName = "cbc-chain"
	
	// DefaultSpecVersion is the expected spec version for CBC Chain
	DefaultSpecVersion = 100
	
	// CBCModules lists all CBC Chain runtime modules
	CBCModules = "System|Timestamp|Balances|TransactionPayment|Sudo|PalletCbcPoi|PalletCbcPos|Dcf"
)

// CBCPalletNames contains the names of CBC-specific pallets
var CBCPalletNames = []string{
	"PalletCbcPoi",  // Proof of Integrity pallet
	"PalletCbcPos",  // Proof of Stake pallet
	"Dcf",           // Deterministic Consensus Framework
}

// IsCBCPallet checks if a pallet name is CBC-specific
func IsCBCPallet(palletName string) bool {
	for _, name := range CBCPalletNames {
		if name == palletName {
			return true
		}
	}
	return false
}

// CBCRuntimeConfig holds CBC Chain runtime configuration
type CBCRuntimeConfig struct {
	// ChainName is the name of the blockchain
	ChainName string
	
	// SpecVersion is the runtime specification version
	SpecVersion int
	
	// EnableDCFFinality indicates if DCF finality verification is enabled
	EnableDCFFinality bool
	
	// BootstrapOnEmpty indicates if runtime should be bootstrapped when DB is empty
	BootstrapOnEmpty bool
	
	// HTTPEndpoint is the HTTP RPC endpoint for metadata fetching
	HTTPEndpoint string
	
	// WSEndpoint is the WebSocket endpoint for subscriptions
	WSEndpoint string
}

// DefaultCBCConfig returns the default CBC Chain configuration
func DefaultCBCConfig(wsEndpoint string) *CBCRuntimeConfig {
	return &CBCRuntimeConfig{
		ChainName:         CBCChainName,
		SpecVersion:       DefaultSpecVersion,
		EnableDCFFinality: true,
		BootstrapOnEmpty:  true,
		HTTPEndpoint:      convertWSToHTTP(wsEndpoint),
		WSEndpoint:        wsEndpoint,
	}
}

// convertWSToHTTP converts a WebSocket endpoint to HTTP
func convertWSToHTTP(wsEndpoint string) string {
	httpEndpoint := wsEndpoint
	if len(httpEndpoint) > 5 && httpEndpoint[:5] == "ws://" {
		httpEndpoint = "http://" + httpEndpoint[5:]
	} else if len(httpEndpoint) > 6 && httpEndpoint[:6] == "wss://" {
		httpEndpoint = "https://" + httpEndpoint[6:]
	}
	return httpEndpoint
}
