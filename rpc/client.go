package rpc

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"go.uber.org/zap"
)

type Client struct {
	url        string
	httpClient *http.Client
	logger     *zap.Logger
}

func NewClient(url string, timeout time.Duration, logger *zap.Logger) *Client {
	return &Client{
		url: url,
		httpClient: &http.Client{
			Timeout: timeout,
		},
		logger: logger,
	}
}

func (c *Client) Forward(ctx context.Context, req *JSONRPCRequest) (*JSONRPCResponse, error) {
	return c.forwardSingle(ctx, req)
}

func (c *Client) ForwardBatch(ctx context.Context, reqs []*JSONRPCRequest) ([]*JSONRPCResponse, error) {
	reqBody, err := json.Marshal(reqs)
	if err != nil {
		c.logger.Error("failed to marshal batch request", zap.Error(err))
		return nil, fmt.Errorf("failed to marshal batch request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", c.url, bytes.NewReader(reqBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create batch request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")

	start := time.Now()
	httpResp, err := c.httpClient.Do(httpReq)
	duration := time.Since(start)

	if err != nil {
		c.logger.Error("upstream batch request failed",
			zap.Error(err),
			zap.Int("batch_size", len(reqs)),
			zap.Duration("duration", duration))
		return nil, fmt.Errorf("upstream batch request failed: %w", err)
	}
	defer httpResp.Body.Close()

	respBody, err := io.ReadAll(httpResp.Body)
	if err != nil {
		c.logger.Error("failed to read batch response", zap.Error(err))
		return nil, fmt.Errorf("failed to read batch response: %w", err)
	}

	if httpResp.StatusCode != http.StatusOK {
		c.logger.Warn("upstream returned non-200 status for batch",
			zap.Int("status", httpResp.StatusCode),
			zap.Int("batch_size", len(reqs)))
		return nil, fmt.Errorf("upstream returned status %d", httpResp.StatusCode)
	}

	var rpcResps []*JSONRPCResponse
	if err := json.Unmarshal(respBody, &rpcResps); err != nil {
		c.logger.Error("failed to unmarshal batch response", zap.Error(err))
		return nil, fmt.Errorf("failed to unmarshal batch response: %w", err)
	}

	c.logger.Debug("batch request forwarded successfully",
		zap.Int("batch_size", len(reqs)),
		zap.Duration("duration", duration),
		zap.Int("status", httpResp.StatusCode))

	return rpcResps, nil
}

func (c *Client) forwardSingle(ctx context.Context, req *JSONRPCRequest) (*JSONRPCResponse, error) {
	reqBody, err := json.Marshal(req)
	if err != nil {
		c.logger.Error("failed to marshal request",
			zap.Error(err),
			zap.String("method", req.Method))
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", c.url, bytes.NewReader(reqBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")

	start := time.Now()
	httpResp, err := c.httpClient.Do(httpReq)
	duration := time.Since(start)

	if err != nil {
		c.logger.Error("upstream request failed",
			zap.Error(err),
			zap.String("method", req.Method),
			zap.Duration("duration", duration))
		return nil, fmt.Errorf("upstream request failed: %w", err)
	}
	defer httpResp.Body.Close()

	respBody, err := io.ReadAll(httpResp.Body)
	if err != nil {
		c.logger.Error("failed to read response",
			zap.Error(err),
			zap.String("method", req.Method))
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if httpResp.StatusCode != http.StatusOK {
		c.logger.Warn("upstream returned non-200 status",
			zap.Int("status", httpResp.StatusCode),
			zap.String("method", req.Method),
			zap.String("body", string(respBody)))

		errMsg := "upstream error"
		if httpResp.StatusCode == http.StatusRequestTimeout || httpResp.StatusCode == http.StatusGatewayTimeout {
			errMsg = "upstream timeout"
		}
		return NewErrorResponse(req.ID, ServerError, errMsg), nil
	}

	var rpcResp JSONRPCResponse
	if err := json.Unmarshal(respBody, &rpcResp); err != nil {
		c.logger.Error("failed to unmarshal response",
			zap.Error(err),
			zap.String("method", req.Method),
			zap.String("body", string(respBody)))
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	c.logger.Debug("request forwarded successfully",
		zap.String("method", req.Method),
		zap.Duration("duration", duration),
		zap.Int("status", httpResp.StatusCode))

	return &rpcResp, nil
}
