package servers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"time"

	cmtbytes "github.com/cometbft/cometbft/libs/bytes"
	cmtcoretypes "github.com/cometbft/cometbft/rpc/core/types"
	cmttypes "github.com/cometbft/cometbft/types"
)

// GenesisDocResponse is a copy of comettypes.GenesisDoc with the exception that InitialHeight is a string. This had to
// be done because for some reason the CometBFT HTTP client expects InitialHeight to be a string even though the
// official type is an int64.
type GenesisDocResponse struct {
	GenesisTime     time.Time                   `json:"genesis_time"`
	ChainID         string                      `json:"chain_id"`
	InitialHeight   string                      `json:"initial_height"`
	ConsensusParams *cmttypes.ConsensusParams   `json:"consensus_params,omitempty"`
	Validators      []cmttypes.GenesisValidator `json:"validators,omitempty"`
	AppHash         cmtbytes.HexBytes           `json:"app_hash"`
	AppState        json.RawMessage             `json:"app_state,omitempty"`
}

// ResultGenesisResponse is a copy of cometcoretypes. ResultGenesis with the exception that it wraps our custom
// GenesisDocResponse type instead of comettypes.GenesisDoc. This had to be done because for some reason the CometBFT
// client expects InitialHeight to be a string even though the official type is an int64.
type ResultGenesisResponse struct {
	Genesis *GenesisDocResponse `json:"genesis"`
}

type MockTendermintServer struct {
	// server is the unerlying HTTP server.
	server *httptest.Server

	// mockGenesis contains the value to be returned by the MockTendermintServer when it receives a Genesis call.
	mockGenesis *ResultGenesisResponse
}

func NewMockTendermintServer() *MockTendermintServer {
	return &MockTendermintServer{}
}

func (m *MockTendermintServer) Start() string {
	handler := http.NewServeMux()
	handler.HandleFunc("/", m.handleRPC)

	m.server = httptest.NewServer(handler)
	return m.server.URL
}

func (m *MockTendermintServer) Stop() {
	if m.server != nil {
		m.server.Close()
	}
}

// SetMockGenesis sets mockGenesis to the value that should be returned by the server when the Genesis call is made.
// This function expects a value of type cometcoretypes.ResultGenesis as parameter instead of our custom type to keep
// integrations with this server as seamless as possible. Any type conversions are done in the function itself instead
// by all integrators.
func (m *MockTendermintServer) SetMockGenesis(mockGenesis *cmtcoretypes.ResultGenesis) {
	m.mockGenesis = &ResultGenesisResponse{
		Genesis: &GenesisDocResponse{
			GenesisTime:     mockGenesis.Genesis.GenesisTime,
			ChainID:         mockGenesis.Genesis.ChainID,
			InitialHeight:   strconv.FormatInt(mockGenesis.Genesis.InitialHeight, 10),
			ConsensusParams: mockGenesis.Genesis.ConsensusParams,
			Validators:      mockGenesis.Genesis.Validators,
			AppHash:         mockGenesis.Genesis.AppHash,
			AppState:        mockGenesis.Genesis.AppState,
		}}
}

func sendJSONRPCResponse(w http.ResponseWriter, id json.RawMessage, result interface{}) {
	resp := struct {
		JSONRPC string          `json:"jsonrpc"`
		ID      json.RawMessage `json:"id"`
		Result  interface{}     `json:"result"`
	}{
		JSONRPC: "2.0",
		ID:      id,
		Result:  result,
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
	}
}

func sendJSONRPCError(w http.ResponseWriter, id json.RawMessage, code int, message string, data interface{}) {
	errorResp := struct {
		JSONRPC string          `json:"jsonrpc"`
		ID      json.RawMessage `json:"id"`
		Error   struct {
			Code    int         `json:"code"`
			Message string      `json:"message"`
			Data    interface{} `json:"data,omitempty"`
		} `json:"error"`
	}{
		JSONRPC: "2.0",
		ID:      id,
		Error: struct {
			Code    int         `json:"code"`
			Message string      `json:"message"`
			Data    interface{} `json:"data,omitempty"`
		}{
			Code:    code,
			Message: message,
			Data:    data,
		},
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(errorResp); err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
	}
}

func (m *MockTendermintServer) handleRPC(w http.ResponseWriter, r *http.Request) {
	var req struct {
		JSONRPC string          `json:"jsonrpc"`
		Method  string          `json:"method"`
		Params  json.RawMessage `json:"params"`
		ID      json.RawMessage `json:"id"`
	}

	// Decode request
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		sendJSONRPCError(w, req.ID, -32700, "Parse error", err)
		return
	}

	// Send response depending on req.Method
	switch req.Method {
	case "genesis":
		sendJSONRPCResponse(w, req.ID, m.mockGenesis)
	default:
		sendJSONRPCError(w, req.ID, -32601, "Method not found", nil)
	}
}
