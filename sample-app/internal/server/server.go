package server

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/securechain-zero/sample-app/internal/config"
)

// Server encapsulates the HTTP server and routing.
type Server struct {
	cfg        config.Config
	httpServer *http.Server
	logger     *slog.Logger
}

// New creates a new Server instance.
func New(cfg config.Config, logger *slog.Logger) *Server {
	s := &Server{
		cfg:    cfg,
		logger: logger,
	}

	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", s.handleHealthz)
	mux.HandleFunc("GET /readyz", s.handleReadyz)
	mux.HandleFunc("GET /info", s.handleInfo)
	mux.HandleFunc("GET /", s.handleRoot)

	// Wrap mux with security and logging middleware
	handler := s.securityHeadersMiddleware(mux)
	handler = s.contentEncodingMiddleware(handler)

	s.httpServer = &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.Port),
		Handler:      handler,
		ReadTimeout:  cfg.ReadTimeout,
		WriteTimeout: cfg.WriteTimeout,
		IdleTimeout:  cfg.IdleTimeout,
	}

	return s
}

// Start begins listening for requests.
func (s *Server) Start() error {
	s.logger.Info("starting HTTP server",
		"port", s.cfg.Port,
		"service", s.cfg.ServiceName,
		"env", s.cfg.Environment,
	)
	if err := s.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		return fmt.Errorf("http server failure: %w", err)
	}
	return nil
}

// Shutdown cleanly stops the server.
func (s *Server) Shutdown(ctx context.Context) error {
	s.logger.Info("shutting down HTTP server gracefully")
	return s.httpServer.Shutdown(ctx)
}

func (s *Server) handleHealthz(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(map[string]string{
		"status": "pass",
		"time":   time.Now().UTC().Format(time.RFC3339),
	})
}

func (s *Server) handleReadyz(w http.ResponseWriter, r *http.Request) {
	// Verify that /tmp is writable (emptyDir memory mount check)
	testFile := fmt.Sprintf("/tmp/readiness-check-%d.tmp", time.Now().UnixNano())
	if err := os.WriteFile(testFile, []byte("ok"), 0600); err != nil {
		s.logger.Error("readiness probe failed: /tmp is not writable", "err", err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusServiceUnavailable)
		_ = json.NewEncoder(w).Encode(map[string]string{
			"status": "fail",
			"error":  "writable /tmp emptyDir requirement not met",
		})
		return
	}
	_ = os.Remove(testFile)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(map[string]string{
		"status": "ready",
		"tmpfs":  "healthy",
	})
}

func (s *Server) handleInfo(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{
		"service":     s.cfg.ServiceName,
		"version":     s.cfg.Version,
		"environment": s.cfg.Environment,
		"standards": []string{
			"SLSA v1.2 Level 3",
			"NIST SP 800-218 (SSDF)",
			"EO 14028",
			"OWASP Top 10 CI/CD Mitigated",
		},
		"security": map[string]any{
			"non_root_uid":        65532,
			"readonly_rootfs":     true,
			"capabilities":        "DROPPED_ALL",
			"compression_brotli":  "DISABLED",
		},
	})
}

func (s *Server) handleRoot(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]string{
		"message": "SecureChain-Zero Hermetic Microservice is operational",
		"status":  "secure",
	})
}

func (s *Server) securityHeadersMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-Frame-Options", "DENY")
		w.Header().Set("Content-Security-Policy", "default-src 'none'")
		w.Header().Set("Referrer-Policy", "no-referrer")
		w.Header().Set("Permissions-Policy", "geolocation=(), microphone=(), camera=()")
		next.ServeHTTP(w, r)
	})
}

// contentEncodingMiddleware enforces identity encoding and drops unsafe compression pathways.
func (s *Server) contentEncodingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Reject any inbound body with br (Brotli) compression
		encoding := r.Header.Get("Content-Encoding")
		if encoding == "br" {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusUnsupportedMediaType)
			_ = json.NewEncoder(w).Encode(map[string]string{
				"error": "Brotli decompression is disabled for security hardening",
			})
			return
		}

		// Enforce identity encoding in response
		w.Header().Set("Accept-Encoding", "identity")
		next.ServeHTTP(w, r)
	})
}
