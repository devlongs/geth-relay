package rpc

import (
	"encoding/json"
	"testing"
)

func TestJSONRPCRequest_Marshal(t *testing.T) {
	req := &JSONRPCRequest{
		JSONRPC: "2.0",
		Method:  "eth_blockNumber",
		Params:  json.RawMessage(`[]`),
		ID:      1,
	}

	data, err := json.Marshal(req)
	if err != nil {
		t.Fatalf("Marshal() failed: %v", err)
	}

	var decoded JSONRPCRequest
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Unmarshal() failed: %v", err)
	}

	if decoded.Method != req.Method {
		t.Errorf("Method = %v, want %v", decoded.Method, req.Method)
	}
}

func TestJSONRPCResponse_Marshal(t *testing.T) {
	resp := &JSONRPCResponse{
		JSONRPC: "2.0",
		Result:  json.RawMessage(`"0x123"`),
		ID:      1,
	}

	data, err := json.Marshal(resp)
	if err != nil {
		t.Fatalf("Marshal() failed: %v", err)
	}

	var decoded JSONRPCResponse
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Unmarshal() failed: %v", err)
	}

	if string(decoded.Result) != string(resp.Result) {
		t.Errorf("Result = %v, want %v", string(decoded.Result), string(resp.Result))
	}
}

func TestNewErrorResponse(t *testing.T) {
	tests := []struct {
		name    string
		id      interface{}
		code    int
		message string
	}{
		{
			name:    "parse error",
			id:      1,
			code:    ParseError,
			message: "parse error",
		},
		{
			name:    "invalid request",
			id:      "test",
			code:    InvalidRequest,
			message: "invalid request",
		},
		{
			name:    "internal error",
			id:      nil,
			code:    InternalError,
			message: "internal error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp := NewErrorResponse(tt.id, tt.code, tt.message)

			if resp.JSONRPC != "2.0" {
				t.Errorf("JSONRPC = %v, want 2.0", resp.JSONRPC)
			}
			if resp.Error == nil {
				t.Fatal("Error is nil")
			}
			if resp.Error.Code != tt.code {
				t.Errorf("Error.Code = %v, want %v", resp.Error.Code, tt.code)
			}
			if resp.Error.Message != tt.message {
				t.Errorf("Error.Message = %v, want %v", resp.Error.Message, tt.message)
			}
			if resp.ID != tt.id {
				t.Errorf("ID = %v, want %v", resp.ID, tt.id)
			}
		})
	}
}

func TestErrorCodes(t *testing.T) {
	codes := map[string]int{
		"ParseError":     ParseError,
		"InvalidRequest": InvalidRequest,
		"MethodNotFound": MethodNotFound,
		"InvalidParams":  InvalidParams,
		"InternalError":  InternalError,
		"ServerError":    ServerError,
	}

	expectedCodes := map[string]int{
		"ParseError":     -32700,
		"InvalidRequest": -32600,
		"MethodNotFound": -32601,
		"InvalidParams":  -32602,
		"InternalError":  -32603,
		"ServerError":    -32000,
	}

	for name, code := range codes {
		if expected, ok := expectedCodes[name]; ok {
			if code != expected {
				t.Errorf("%s = %d, want %d", name, code, expected)
			}
		}
	}
}
