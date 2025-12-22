package rpc

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"go.uber.org/zap"
)

func TestNewClient(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	client := NewClient("http://localhost:8546", 30*time.Second, logger)

	if client == nil {
		t.Fatal("NewClient() returned nil")
	}
	if client.url != "http://localhost:8546" {
		t.Errorf("url = %v, want http://localhost:8546", client.url)
	}
}

func TestClient_Forward(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := JSONRPCResponse{
			JSONRPC: "2.0",
			Result:  json.RawMessage(`"0x123"`),
			ID:      1,
		}
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	logger, _ := zap.NewDevelopment()
	client := NewClient(server.URL, 5*time.Second, logger)

	req := &JSONRPCRequest{
		JSONRPC: "2.0",
		Method:  "eth_blockNumber",
		Params:  json.RawMessage(`[]`),
		ID:      1,
	}

	resp, err := client.Forward(context.Background(), req)
	if err != nil {
		t.Fatalf("Forward() error = %v", err)
	}

	if resp.JSONRPC != "2.0" {
		t.Errorf("JSONRPC = %v, want 2.0", resp.JSONRPC)
	}
	if string(resp.Result) != `"0x123"` {
		t.Errorf("Result = %v, want \"0x123\"", string(resp.Result))
	}
}

func TestClient_ForwardBatch(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := []*JSONRPCResponse{
			{JSONRPC: "2.0", Result: json.RawMessage(`"0x1"`), ID: 1},
			{JSONRPC: "2.0", Result: json.RawMessage(`"0x2"`), ID: 2},
		}
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	logger, _ := zap.NewDevelopment()
	client := NewClient(server.URL, 5*time.Second, logger)

	reqs := []*JSONRPCRequest{
		{JSONRPC: "2.0", Method: "eth_blockNumber", Params: json.RawMessage(`[]`), ID: 1},
		{JSONRPC: "2.0", Method: "eth_gasPrice", Params: json.RawMessage(`[]`), ID: 2},
	}

	resps, err := client.ForwardBatch(context.Background(), reqs)
	if err != nil {
		t.Fatalf("ForwardBatch() error = %v", err)
	}

	if len(resps) != 2 {
		t.Errorf("len(resps) = %d, want 2", len(resps))
	}
}

func TestClient_ForwardError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("server error"))
	}))
	defer server.Close()

	logger, _ := zap.NewDevelopment()
	client := NewClient(server.URL, 5*time.Second, logger)

	req := &JSONRPCRequest{
		JSONRPC: "2.0",
		Method:  "eth_blockNumber",
		Params:  json.RawMessage(`[]`),
		ID:      1,
	}

	resp, err := client.Forward(context.Background(), req)
	if err != nil {
		t.Fatalf("Forward() error = %v", err)
	}

	if resp.Error == nil {
		t.Error("Expected error response, got nil")
	}
}
