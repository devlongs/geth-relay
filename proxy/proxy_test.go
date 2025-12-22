package proxy

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/devlongs/geth-relay/rpc"
	"go.uber.org/zap"
)

func TestNew(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	client := rpc.NewClient("http://localhost:8546", 30*time.Second, logger)
	proxy := New(client, logger, 100, 25000000)

	if proxy == nil {
		t.Fatal("New() returned nil")
	}
	if proxy.maxBatchItems != 100 {
		t.Errorf("maxBatchItems = %d, want 100", proxy.maxBatchItems)
	}
}

func TestProxy_HandleRequest(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := rpc.JSONRPCResponse{
			JSONRPC: "2.0",
			Result:  json.RawMessage(`"0x123"`),
			ID:      1,
		}
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	logger, _ := zap.NewDevelopment()
	client := rpc.NewClient(server.URL, 5*time.Second, logger)
	proxy := New(client, logger, 100, 25000000)

	req := &rpc.JSONRPCRequest{
		JSONRPC: "2.0",
		Method:  "eth_blockNumber",
		Params:  json.RawMessage(`[]`),
		ID:      1,
	}

	resp := proxy.HandleRequest(context.Background(), req)
	if resp.Error != nil {
		t.Errorf("HandleRequest() error = %v", resp.Error)
	}
	if string(resp.Result) != `"0x123"` {
		t.Errorf("Result = %v, want \"0x123\"", string(resp.Result))
	}
}

func TestProxy_HandleRequestInvalidVersion(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	client := rpc.NewClient("http://localhost:8546", 30*time.Second, logger)
	proxy := New(client, logger, 100, 25000000)

	req := &rpc.JSONRPCRequest{
		JSONRPC: "1.0",
		Method:  "eth_blockNumber",
		ID:      1,
	}

	resp := proxy.HandleRequest(context.Background(), req)
	if resp.Error == nil {
		t.Error("Expected error for invalid JSONRPC version")
	}
	if resp.Error.Code != rpc.InvalidRequest {
		t.Errorf("Error code = %d, want %d", resp.Error.Code, rpc.InvalidRequest)
	}
}

func TestProxy_HandleRequestEmptyMethod(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	client := rpc.NewClient("http://localhost:8546", 30*time.Second, logger)
	proxy := New(client, logger, 100, 25000000)

	req := &rpc.JSONRPCRequest{
		JSONRPC: "2.0",
		Method:  "",
		ID:      1,
	}

	resp := proxy.HandleRequest(context.Background(), req)
	if resp.Error == nil {
		t.Error("Expected error for empty method")
	}
}

func TestProxy_HandleBatchRequest(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := []*rpc.JSONRPCResponse{
			{JSONRPC: "2.0", Result: json.RawMessage(`"0x1"`), ID: 1},
			{JSONRPC: "2.0", Result: json.RawMessage(`"0x2"`), ID: 2},
		}
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	logger, _ := zap.NewDevelopment()
	client := rpc.NewClient(server.URL, 5*time.Second, logger)
	proxy := New(client, logger, 100, 25000000)

	reqs := []*rpc.JSONRPCRequest{
		{JSONRPC: "2.0", Method: "eth_blockNumber", Params: json.RawMessage(`[]`), ID: 1},
		{JSONRPC: "2.0", Method: "eth_gasPrice", Params: json.RawMessage(`[]`), ID: 2},
	}

	resps := proxy.HandleBatchRequest(context.Background(), reqs)
	if len(resps) != 2 {
		t.Errorf("len(resps) = %d, want 2", len(resps))
	}
}

func TestProxy_HandleBatchRequestExceedsLimit(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	client := rpc.NewClient("http://localhost:8546", 30*time.Second, logger)
	proxy := New(client, logger, 2, 25000000)

	reqs := []*rpc.JSONRPCRequest{
		{JSONRPC: "2.0", Method: "eth_blockNumber", ID: 1},
		{JSONRPC: "2.0", Method: "eth_gasPrice", ID: 2},
		{JSONRPC: "2.0", Method: "eth_chainId", ID: 3},
	}

	resps := proxy.HandleBatchRequest(context.Background(), reqs)
	if len(resps) != 1 {
		t.Errorf("len(resps) = %d, want 1", len(resps))
	}
	if resps[0].Error == nil {
		t.Error("Expected error for batch exceeding limit")
	}
}

func TestProxy_HandleBatchRequestEmpty(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	client := rpc.NewClient("http://localhost:8546", 30*time.Second, logger)
	proxy := New(client, logger, 100, 25000000)

	resps := proxy.HandleBatchRequest(context.Background(), []*rpc.JSONRPCRequest{})
	if len(resps) != 1 {
		t.Errorf("len(resps) = %d, want 1", len(resps))
	}
	if resps[0].Error == nil {
		t.Error("Expected error for empty batch")
	}
}
