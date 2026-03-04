package web

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/dave1010/jorin/internal/agent"
	"github.com/dave1010/jorin/internal/prompt"
	"github.com/dave1010/jorin/internal/types"
)

const defaultAddr = "127.0.0.1:8080"

// Config configures the web server mode.
type Config struct {
	Addr   string
	Model  string
	Agent  agent.Agent
	Policy *types.Policy
	ErrOut io.Writer
}

// Server serves the browser UI and JSON chat API.
type Server struct {
	cfg Config
}

// NewServer creates a web server with sensible defaults.
func NewServer(cfg Config) *Server {
	if strings.TrimSpace(cfg.Addr) == "" {
		cfg.Addr = defaultAddr
	}
	return &Server{cfg: cfg}
}

// Run starts the HTTP server and shuts down cleanly when the context is canceled.
func (s *Server) Run(ctx context.Context) error {
	if s.cfg.Agent == nil {
		return errors.New("web server requires an agent")
	}
	handler := s.routes()
	srv := &http.Server{
		Addr:              s.cfg.Addr,
		Handler:           handler,
		ReadHeaderTimeout: 5 * time.Second,
	}

	errCh := make(chan error, 1)
	go func() {
		if s.cfg.ErrOut != nil {
			_, _ = fmt.Fprintf(s.cfg.ErrOut, "Jorin web UI listening on http://%s\n", s.cfg.Addr)
		}
		err := srv.ListenAndServe()
		if !errors.Is(err, http.ErrServerClosed) {
			errCh <- err
			return
		}
		errCh <- nil
	}()

	select {
	case <-ctx.Done():
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = srv.Shutdown(shutdownCtx)
		<-errCh
		return nil
	case err := <-errCh:
		return err
	}
}

func (s *Server) routes() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/", s.handleIndex)
	mux.HandleFunc("/api/chat", s.handleChat)
	mux.HandleFunc("/healthz", s.handleHealth)
	return mux
}

type chatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type chatRequest struct {
	Messages []chatMessage `json:"messages"`
	Prompt   string        `json:"prompt"`
}

type chatResponse struct {
	Reply string `json:"reply"`
}

func (s *Server) handleChat(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req chatRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid JSON body", http.StatusBadRequest)
		return
	}

	msgs := []types.Message{{Role: "system", Content: prompt.SystemPrompt()}}
	for _, msg := range req.Messages {
		role := strings.TrimSpace(msg.Role)
		if role != "user" && role != "assistant" {
			http.Error(w, "messages role must be user or assistant", http.StatusBadRequest)
			return
		}
		if strings.TrimSpace(msg.Content) == "" {
			continue
		}
		msgs = append(msgs, types.Message{Role: role, Content: msg.Content})
	}
	if p := strings.TrimSpace(req.Prompt); p != "" {
		msgs = append(msgs, types.Message{Role: "user", Content: req.Prompt})
	}
	if len(msgs) < 2 {
		http.Error(w, "prompt is required", http.StatusBadRequest)
		return
	}

	_, out, err := s.cfg.Agent.ChatSession(s.cfg.Model, msgs, s.cfg.Policy)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadGateway)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(chatResponse{Reply: out}); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	_, _ = io.WriteString(w, "ok")
}

func (s *Server) handleIndex(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_, _ = io.WriteString(w, indexHTML)
}
