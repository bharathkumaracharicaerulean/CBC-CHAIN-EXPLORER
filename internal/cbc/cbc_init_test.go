package cbc

import (
	"testing"
)

func TestNewCBCInitializer(t *testing.T) {
	tests := []struct {
		name        string
		wsEndpoint  string
		wantHTTP    string
	}{
		{
			name:       "WebSocket to HTTP conversion",
			wsEndpoint: "ws://172.17.0.1:9933",
			wantHTTP:   "http://172.17.0.1:9933",
		},
		{
			name:       "Secure WebSocket to HTTPS conversion",
			wsEndpoint: "wss://example.com:9944",
			wantHTTP:   "https://example.com:9944",
		},
		{
			name:       "Already HTTP endpoint",
			wsEndpoint: "http://localhost:9933",
			wantHTTP:   "http://localhost:9933",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			init := NewCBCInitializer(nil, tt.wsEndpoint)
			if init.httpURL != tt.wantHTTP {
				t.Errorf("NewCBCInitializer() httpURL = %v, want %v", init.httpURL, tt.wantHTTP)
			}
		})
	}
}

func TestIsCBCPallet(t *testing.T) {
	tests := []struct {
		name       string
		palletName string
		want       bool
	}{
		{
			name:       "CBC PoI pallet",
			palletName: "PalletCbcPoi",
			want:       true,
		},
		{
			name:       "CBC PoS pallet",
			palletName: "PalletCbcPos",
			want:       true,
		},
		{
			name:       "DCF pallet",
			palletName: "Dcf",
			want:       true,
		},
		{
			name:       "Standard System pallet",
			palletName: "System",
			want:       false,
		},
		{
			name:       "Standard Balances pallet",
			palletName: "Balances",
			want:       false,
		},
		{
			name:       "Non-existent pallet",
			palletName: "FakePallet",
			want:       false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsCBCPallet(tt.palletName); got != tt.want {
				t.Errorf("IsCBCPallet() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDefaultCBCConfig(t *testing.T) {
	wsEndpoint := "ws://172.17.0.1:9933"
	config := DefaultCBCConfig(wsEndpoint)

	if config.ChainName != CBCChainName {
		t.Errorf("DefaultCBCConfig() ChainName = %v, want %v", config.ChainName, CBCChainName)
	}

	if config.SpecVersion != DefaultSpecVersion {
		t.Errorf("DefaultCBCConfig() SpecVersion = %v, want %v", config.SpecVersion, DefaultSpecVersion)
	}

	if !config.EnableDCFFinality {
		t.Error("DefaultCBCConfig() EnableDCFFinality should be true")
	}

	if !config.BootstrapOnEmpty {
		t.Error("DefaultCBCConfig() BootstrapOnEmpty should be true")
	}

	expectedHTTP := "http://172.17.0.1:9933"
	if config.HTTPEndpoint != expectedHTTP {
		t.Errorf("DefaultCBCConfig() HTTPEndpoint = %v, want %v", config.HTTPEndpoint, expectedHTTP)
	}

	if config.WSEndpoint != wsEndpoint {
		t.Errorf("DefaultCBCConfig() WSEndpoint = %v, want %v", config.WSEndpoint, wsEndpoint)
	}
}

func TestConvertWSToHTTP(t *testing.T) {
	tests := []struct {
		name       string
		wsEndpoint string
		want       string
	}{
		{
			name:       "ws:// to http://",
			wsEndpoint: "ws://localhost:9933",
			want:       "http://localhost:9933",
		},
		{
			name:       "wss:// to https://",
			wsEndpoint: "wss://example.com:9944",
			want:       "https://example.com:9944",
		},
		{
			name:       "Already http://",
			wsEndpoint: "http://localhost:9933",
			want:       "http://localhost:9933",
		},
		{
			name:       "Already https://",
			wsEndpoint: "https://example.com:9944",
			want:       "https://example.com:9944",
		},
		{
			name:       "Docker host gateway",
			wsEndpoint: "ws://172.17.0.1:9933",
			want:       "http://172.17.0.1:9933",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := convertWSToHTTP(tt.wsEndpoint); got != tt.want {
				t.Errorf("convertWSToHTTP() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRPCRequestStructure(t *testing.T) {
	req := RPCRequest{
		ID:      1,
		JSONRPC: "2.0",
		Method:  "state_getMetadata",
		Params:  []interface{}{},
	}

	if req.ID != 1 {
		t.Errorf("RPCRequest ID = %v, want 1", req.ID)
	}

	if req.JSONRPC != "2.0" {
		t.Errorf("RPCRequest JSONRPC = %v, want 2.0", req.JSONRPC)
	}

	if req.Method != "state_getMetadata" {
		t.Errorf("RPCRequest Method = %v, want state_getMetadata", req.Method)
	}
}

func TestCBCModulesConstant(t *testing.T) {
	expectedModules := "System|Timestamp|Balances|TransactionPayment|Sudo|PalletCbcPoi|PalletCbcPos|Dcf"
	if CBCModules != expectedModules {
		t.Errorf("CBCModules = %v, want %v", CBCModules, expectedModules)
	}
}

func TestCBCPalletNames(t *testing.T) {
	expectedPallets := []string{"PalletCbcPoi", "PalletCbcPos", "Dcf"}
	
	if len(CBCPalletNames) != len(expectedPallets) {
		t.Errorf("CBCPalletNames length = %v, want %v", len(CBCPalletNames), len(expectedPallets))
	}

	for i, name := range expectedPallets {
		if CBCPalletNames[i] != name {
			t.Errorf("CBCPalletNames[%d] = %v, want %v", i, CBCPalletNames[i], name)
		}
	}
}
