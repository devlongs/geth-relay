package proxy

import (
	"context"

	"github.com/devlongs/geth-relay/rpc"
	"go.uber.org/zap"
)

type Proxy struct {
	client        *rpc.Client
	logger        *zap.Logger
	maxBatchItems int
	maxBatchSize  int
}

func New(client *rpc.Client, logger *zap.Logger, maxBatchItems, maxBatchSize int) *Proxy {
	return &Proxy{
		client:        client,
		logger:        logger,
		maxBatchItems: maxBatchItems,
		maxBatchSize:  maxBatchSize,
	}
}

func (p *Proxy) HandleRequest(ctx context.Context, req *rpc.JSONRPCRequest) *rpc.JSONRPCResponse {
	if req.JSONRPC != "2.0" {
		p.logger.Warn("invalid jsonrpc version",
			zap.String("version", req.JSONRPC),
			zap.String("method", req.Method))
		return rpc.NewErrorResponse(req.ID, rpc.InvalidRequest, "jsonrpc must be 2.0")
	}

	if req.Method == "" {
		p.logger.Warn("empty method in request")
		return rpc.NewErrorResponse(req.ID, rpc.InvalidRequest, "method cannot be empty")
	}

	p.logger.Info("handling request",
		zap.String("method", req.Method),
		zap.Any("id", req.ID))

	resp, err := p.client.Forward(ctx, req)
	if err != nil {
		p.logger.Error("failed to forward request",
			zap.Error(err),
			zap.String("method", req.Method))
		return rpc.NewErrorResponse(req.ID, rpc.InternalError, "failed to forward request to upstream")
	}

	return resp
}

func (p *Proxy) HandleBatchRequest(ctx context.Context, reqs []*rpc.JSONRPCRequest) []*rpc.JSONRPCResponse {
	if len(reqs) == 0 {
		p.logger.Warn("empty batch request")
		return []*rpc.JSONRPCResponse{
			rpc.NewErrorResponse(nil, rpc.InvalidRequest, "empty batch request"),
		}
	}

	if len(reqs) > p.maxBatchItems {
		p.logger.Warn("batch request exceeds limit",
			zap.Int("size", len(reqs)),
			zap.Int("limit", p.maxBatchItems))
		return []*rpc.JSONRPCResponse{
			{
				JSONRPC: "2.0",
				Error:   rpc.ErrBatchTooLarge,
				ID:      nil,
			},
		}
	}

	p.logger.Info("handling batch request",
		zap.Int("size", len(reqs)))

	for i, req := range reqs {
		if req.JSONRPC != "2.0" {
			p.logger.Warn("invalid jsonrpc version in batch",
				zap.Int("index", i),
				zap.String("version", req.JSONRPC))
			return []*rpc.JSONRPCResponse{
				rpc.NewErrorResponse(req.ID, rpc.InvalidRequest, "jsonrpc must be 2.0"),
			}
		}
	}

	resps, err := p.client.ForwardBatch(ctx, reqs)
	if err != nil {
		p.logger.Error("failed to forward batch request",
			zap.Error(err),
			zap.Int("size", len(reqs)))

		errorResps := make([]*rpc.JSONRPCResponse, len(reqs))
		for i, req := range reqs {
			errorResps[i] = rpc.NewErrorResponse(req.ID, rpc.InternalError, "failed to forward batch request")
		}
		return errorResps
	}

	return resps
}
