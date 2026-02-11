package httpserver

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"greenapi-form/internal/greenapi"
)

type Server struct {
	addr   string
	mux    *http.ServeMux
	client *greenapi.Client
}

func New(addr, greenAPIBaseURL string) (*Server, error) {
	if strings.TrimSpace(addr) == "" {
		return nil, errors.New("server address is required")
	}

	client, err := greenapi.New(greenAPIBaseURL)
	if err != nil {
		return nil, fmt.Errorf("init green api client: %w", err)
	}

	s := &Server{
		addr:   addr,
		mux:    http.NewServeMux(),
		client: client,
	}

	s.routes()
	return s, nil
}

func (s *Server) Addr() string {
	return s.addr
}

func (s *Server) Run() error {
	server := &http.Server{
		Addr:    s.addr,
		Handler: s.withCORS(s.mux),
	}

	if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		return fmt.Errorf("listen and serve: %w", err)
	}

	return nil
}

func (s *Server) routes() {
	fileServer := http.FileServer(http.Dir("web"))
	s.mux.Handle("/assets/", http.StripPrefix("/", fileServer))
	s.mux.Handle("/wasm_exec.js", http.StripPrefix("/", fileServer))
	s.mux.Handle("/", http.HandlerFunc(s.handleIndex))
	s.mux.Handle("/api/call", http.HandlerFunc(s.handleCall))
}

func (s *Server) handleIndex(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	http.ServeFile(w, r, "web/index.html")
}

type callRequest struct {
	IDInstance       string         `json:"idInstance"`
	APITokenInstance string         `json:"apiTokenInstance"`
	Method           string         `json:"method"`
	Payload          map[string]any `json:"payload"`
}

type callResponse struct {
	Result json.RawMessage `json:"result,omitempty"`
	Error  string          `json:"error,omitempty"`
}

func (s *Server) handleCall(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	defer r.Body.Close()

	var req callRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, callResponse{Error: "invalid JSON body"})
		return
	}

	req.IDInstance = strings.TrimSpace(req.IDInstance)
	req.APITokenInstance = strings.TrimSpace(req.APITokenInstance)
	req.Method = strings.TrimSpace(req.Method)

	if req.IDInstance == "" || req.APITokenInstance == "" || req.Method == "" {
		writeJSON(w, http.StatusBadRequest, callResponse{Error: "idInstance, apiTokenInstance and method are required"})
		return
	}

	resp, err := s.client.Call(context.Background(), req.IDInstance, req.APITokenInstance, req.Method, req.Payload)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, callResponse{Error: err.Error()})
		return
	}

	writeJSON(w, http.StatusOK, callResponse{Result: resp})
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}

func (s *Server) withCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

		next.ServeHTTP(w, r)
	})
}
