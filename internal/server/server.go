package server

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/devlongs/geth-relay/proxy"
	"github.com/devlongs/geth-relay/rpc"
	"go.uber.org/zap"
)

type Server struct {
	proxy       *proxy.Proxy
	logger      *zap.Logger
	httpServer  *http.Server
	maxBodySize int
}

func New(addr string, p *proxy.Proxy, logger *zap.Logger, maxBodySize int) *Server {
	s := &Server{
		proxy:       p,
		logger:      logger,
		maxBodySize: maxBodySize,
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/", s.RPCHandler)
	mux.HandleFunc("/health", s.HealthHandler)

	handler := RecoveryMiddleware(logger)(LoggingMiddleware(logger)(mux))

	s.httpServer = &http.Server{
		Addr:         addr,
		Handler:      handler,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	return s
}

func (s *Server) Start() error {
	s.logger.Info("starting server", zap.String("addr", s.httpServer.Addr))
	if err := s.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		return fmt.Errorf("failed to start server: %w", err)
	}
	return nil
}

func (s *Server) Shutdown(ctx context.Context) error {
	s.logger.Info("shutting down server")
	return s.httpServer.Shutdown(ctx)
}

func (s *Server) RPCHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		s.logger.Warn("invalid method", zap.String("method", r.Method))
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if r.ContentLength > int64(s.maxBodySize) {
		s.logger.Warn("request body too large",
			zap.Int64("content_length", r.ContentLength),
			zap.Int("max_size", s.maxBodySize))
		s.writeErrorResponse(w, nil, rpc.InvalidRequest, "request body too large")
		return
	}

	body, err := io.ReadAll(io.LimitReader(r.Body, int64(s.maxBodySize)))
	if err != nil {
		s.logger.Error("failed to read request body", zap.Error(err))
		s.writeErrorResponse(w, nil, rpc.ParseError, "failed to read request")
		return
	}
	defer r.Body.Close()

	body = bytes.TrimSpace(body)

	isBatch := len(body) > 0 && body[0] == '['

	if isBatch {
		var batchReqs []*rpc.JSONRPCRequest
		if err := json.Unmarshal(body, &batchReqs); err != nil {
			s.logger.Error("failed to unmarshal batch request", zap.Error(err))
			s.writeErrorResponse(w, nil, rpc.ParseError, "invalid json")
			return
		}

		batchResps := s.proxy.HandleBatchRequest(r.Context(), batchReqs)

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(batchResps); err != nil {
			s.logger.Error("failed to encode batch response", zap.Error(err))
		}
	} else {
		var req rpc.JSONRPCRequest
		if err := json.Unmarshal(body, &req); err != nil {
			s.logger.Error("failed to unmarshal request", zap.Error(err), zap.String("body", string(body)))
			s.writeErrorResponse(w, nil, rpc.ParseError, "invalid json")
			return
		}

		resp := s.proxy.HandleRequest(r.Context(), &req)

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(resp); err != nil {
			s.logger.Error("failed to encode response", zap.Error(err))
		}
	}
}

func (s *Server) writeErrorResponse(w http.ResponseWriter, id interface{}, code int, message string) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(rpc.NewErrorResponse(id, code, message))
}
