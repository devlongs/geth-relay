package server

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/devlongs/geth-relay/proxy"
	"github.com/devlongs/geth-relay/rpc"
	"go.uber.org/zap"
)

func setupTestServer() *Server {
	logger, _ := zap.NewDevelopment()
	client := rpc.NewClient("http://localhost:8546", 30*time.Second, logger)
	proxyHandler := proxy.New(client, logger, 100, 25000000)
	return New("localhost:8545", proxyHandler, logger, 5242880)
}

func TestNew(t *testing.T) {
	server := setupTestServer()
	if server == nil {
		t.Fatal("New() returned nil")
	}
	if server.httpServer == nil {
		t.Error("httpServer is nil")
	}
}

func TestServer_HealthHandler(t *testing.T) {
	server := setupTestServer()

	req := httptest.NewRequest("GET", "/health", nil)
	w := httptest.NewRecorder()

	server.HealthHandler(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("status code = %d, want %d", w.Code, http.StatusOK)
	}

	var resp map[string]string
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if resp["status"] != "healthy" {
		t.Errorf("status = %v, want healthy", resp["status"])
	}
}

func TestServer_RPCHandlerInvalidMethod(t *testing.T) {
	server := setupTestServer()

	req := httptest.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()

	server.RPCHandler(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("status code = %d, want %d", w.Code, http.StatusMethodNotAllowed)
	}
}

func TestServer_RPCHandlerInvalidJSON(t *testing.T) {
	server := setupTestServer()

	req := httptest.NewRequest("POST", "/", bytes.NewBufferString("invalid json"))
	w := httptest.NewRecorder()

	server.RPCHandler(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("status code = %d, want %d", w.Code, http.StatusOK)
	}

	var resp rpc.JSONRPCResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if resp.Error == nil {
		t.Error("Expected error response for invalid JSON")
	}
	if resp.Error.Code != rpc.ParseError {
		t.Errorf("error code = %d, want %d", resp.Error.Code, rpc.ParseError)
	}
}

func TestServer_RPCHandlerRequestTooLarge(t *testing.T) {
	server := setupTestServer()

	largeBody := bytes.Repeat([]byte("x"), 6*1024*1024)
	req := httptest.NewRequest("POST", "/", bytes.NewReader(largeBody))
	req.ContentLength = int64(len(largeBody))
	w := httptest.NewRecorder()

	server.RPCHandler(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("status code = %d, want %d", w.Code, http.StatusOK)
	}
}

func TestServer_RPCHandlerBatchRequest(t *testing.T) {
	server := setupTestServer()

	batch := []map[string]interface{}{
		{"jsonrpc": "2.0", "method": "eth_blockNumber", "params": []interface{}{}, "id": 1},
		{"jsonrpc": "2.0", "method": "eth_gasPrice", "params": []interface{}{}, "id": 2},
	}
	body, _ := json.Marshal(batch)

	req := httptest.NewRequest("POST", "/", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	server.RPCHandler(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("status code = %d, want %d", w.Code, http.StatusOK)
	}
}
