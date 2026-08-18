package api

import (
	"encoding/json"
	"errors"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"

	"tiny-gpu-bench/internal/bench"
	"tiny-gpu-bench/internal/origin"
)

// Server serves HTTP API and static SPA assets.
type Server struct {
	bench   *bench.Bench
	webRoot string
	mux     *http.ServeMux
}

// NewServer creates an API server.
func NewServer(b *bench.Bench, webRoot string) *Server {
	s := &Server{bench: b, webRoot: webRoot, mux: http.NewServeMux()}
	s.routes()
	return s
}

func (s *Server) Handler() http.Handler {
	return corsMiddleware(s.mux)
}

func (s *Server) routes() {
	s.mux.HandleFunc("GET /health", s.handleHealth)
	s.mux.HandleFunc("GET /api/status", s.handleStatus)
	s.mux.HandleFunc("POST /api/simulate", s.handleSimulate)
	s.mux.HandleFunc("POST /api/step", s.handleStep)
	s.mux.HandleFunc("POST /api/reset", s.handleReset)
	s.mux.HandleFunc("POST /api/button", s.handleButton)
	s.mux.HandleFunc("GET /api/files", s.handleListFiles)
	s.mux.HandleFunc("GET /api/files/{path...}", s.handleReadFile)
	s.mux.HandleFunc("PUT /api/files/{path...}", s.handleWriteFile)
	s.mux.HandleFunc("POST /api/gate-count", s.handleGateCount)
	s.mux.HandleFunc("POST /api/export-fpga", s.handleExportFPGA)
	s.mux.HandleFunc("POST /api/load-template", s.handleLoadTemplate)
	s.mux.HandleFunc("GET /api/uart", s.handleUART)
	s.mux.HandleFunc("GET /api/leds", s.handleLEDs)
	s.mux.HandleFunc("GET /api/servos", s.handleServos)
	s.mux.HandleFunc("GET /api/waves", s.handleWaves)
	s.mux.HandleFunc("GET /api/framebuffer", s.handleFramebuffer)
	s.mux.HandleFunc("GET /api/framebuffer.png", s.handleFramebufferPNG)

	if s.webRoot != "" {
		s.mux.Handle("/", spaHandler(s.webRoot))
	}
}

func (s *Server) handleHealth(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]bool{"ok": true})
}

func (s *Server) handleStatus(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, s.bench.Status())
}

func (s *Server) handleSimulate(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := requestContext(r, 120*time.Second)
	defer cancel()
	res := s.bench.Simulate(ctx)
	writeSimResult(w, res)
}

func (s *Server) handleStep(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := requestContext(r, 30*time.Second)
	defer cancel()
	res := s.bench.Step(ctx)
	writeSimResult(w, res)
}

func (s *Server) handleReset(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := requestContext(r, 30*time.Second)
	defer cancel()
	res := s.bench.Reset(ctx)
	writeSimResult(w, res)
}

func (s *Server) handleButton(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Down bool `json:"down"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid json"})
		return
	}
	if err := s.bench.SetButton(body.Down); err != nil {
		status := http.StatusBadRequest
		if strings.HasPrefix(err.Error(), "busy") {
			status = http.StatusConflict
		}
		writeJSON(w, status, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, map[string]bool{"ok": true})
}

func (s *Server) handleListFiles(w http.ResponseWriter, _ *http.Request) {
	files, err := s.bench.Workspace().List()
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"files": files})
}

func (s *Server) handleReadFile(w http.ResponseWriter, r *http.Request) {
	path := r.PathValue("path")
	content, err := s.bench.Workspace().Read(path)
	if err != nil {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"path": path, "content": content})
}

func (s *Server) handleWriteFile(w http.ResponseWriter, r *http.Request) {
	path := r.PathValue("path")
	var body struct {
		Content string `json:"content"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid json"})
		return
	}
	if err := s.bench.Workspace().Write(path, body.Content); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, map[string]bool{"ok": true})
}

func (s *Server) handleGateCount(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := requestContext(r, 120*time.Second)
	defer cancel()
	res := s.bench.GateCount(ctx)
	if res.Error == "busy" {
		writeJSON(w, http.StatusConflict, res)
		return
	}
	status := http.StatusOK
	if !res.OK {
		status = http.StatusInternalServerError
	}
	writeJSON(w, status, res)
}

func (s *Server) handleExportFPGA(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := requestContext(r, 300*time.Second)
	defer cancel()
	data, log, err := s.bench.ExportFPGA(ctx)
	if err != nil {
		if err.Error() == "busy" {
			writeJSON(w, http.StatusConflict, map[string]string{"error": "busy", "log": log})
			return
		}
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error(), "log": log})
		return
	}
	w.Header().Set("Content-Type", "application/octet-stream")
	w.Header().Set("Content-Disposition", `attachment; filename="tiny-gpu-bench.bin"`)
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(data)
}

func (s *Server) handleLoadTemplate(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Template string `json:"template"`
	}
	_ = json.NewDecoder(r.Body).Decode(&body)
	name := body.Template
	if name == "" {
		name = "hello-gpu"
	}
	if err := s.bench.LoadTemplate(name); err != nil {
		if err.Error() == "busy" {
			writeJSON(w, http.StatusConflict, map[string]string{"error": "busy"})
			return
		}
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, map[string]bool{"ok": true})
}

func (s *Server) handleUART(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"text": s.bench.UART()})
}

func (s *Server) handleLEDs(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]int{"leds": s.bench.LEDs()})
}

func (s *Server) handleServos(w http.ResponseWriter, _ *http.Request) {
	d := s.bench.Servos()
	writeJSON(w, http.StatusOK, map[string][4]int{"duty": d})
}

func (s *Server) handleWaves(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, s.bench.Waves())
}

func (s *Server) handleFramebuffer(w http.ResponseWriter, _ *http.Request) {
	fb := s.bench.Framebuffer()
	if len(fb) == 0 {
		fb = make([]byte, 4096)
	}
	w.Header().Set("Content-Type", "application/octet-stream")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(fb)
}

func (s *Server) handleFramebufferPNG(w http.ResponseWriter, _ *http.Request) {
	png, err := framebufferPNG(s.bench.Framebuffer())
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	w.Header().Set("Content-Type", "image/png")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(png)
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		reqOrigin := r.Header.Get("Origin")
		if reqOrigin != "" {
			if !origin.Allow(reqOrigin) {
				http.Error(w, "origin not allowed", http.StatusForbidden)
				return
			}
			w.Header().Set("Access-Control-Allow-Origin", reqOrigin)
			w.Header().Set("Vary", "Origin")
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		}
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func writeSimResult(w http.ResponseWriter, res bench.SimulateResult) {
	if res.Error == "busy" {
		writeJSON(w, http.StatusConflict, res)
		return
	}
	status := http.StatusOK
	if !res.OK {
		status = http.StatusInternalServerError
	}
	writeJSON(w, status, res)
}

func spaHandler(webRoot string) http.Handler {
	root, err := filepath.Abs(webRoot)
	if err != nil {
		root = webRoot
	}
	index := filepath.Join(root, "index.html")
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.Path, "/api/") || r.URL.Path == "/health" || r.URL.Path == "/mcp" {
			http.NotFound(w, r)
			return
		}
		rel := strings.TrimPrefix(path.Clean("/"+r.URL.Path), "/")
		if rel == "." || rel == "" || strings.HasSuffix(r.URL.Path, "/") {
			http.ServeFile(w, r, index)
			return
		}
		abs := filepath.Join(root, filepath.FromSlash(rel))
		if abs != root && !strings.HasPrefix(abs, root+string(filepath.Separator)) {
			http.ServeFile(w, r, index)
			return
		}
		st, err := os.Stat(abs)
		if err != nil || st.IsDir() {
			if errors.Is(err, os.ErrNotExist) || (err == nil && st.IsDir()) {
				http.ServeFile(w, r, index)
				return
			}
			http.NotFound(w, r)
			return
		}
		http.ServeFile(w, r, abs)
	})
}
