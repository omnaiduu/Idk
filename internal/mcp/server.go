package mcp

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"tiny-gpu-bench/internal/bench"
	"tiny-gpu-bench/internal/origin"
)

// Server exposes MCP tools backed by the shared bench service.
type Server struct {
	bench *bench.Bench
}

// NewServer registers all MCP tools.
func NewServer(b *bench.Bench) *Server {
	return &Server{bench: b}
}

// Handler returns the streamable HTTP handler with origin checks.
func (s *Server) Handler() http.Handler {
	mcpServer := mcp.NewServer(&mcp.Implementation{
		Name:    "tiny-gpu-bench",
		Version: "1.0.0",
	}, nil)

	registerTools(mcpServer, s)
	base := mcp.NewStreamableHTTPHandler(func(*http.Request) *mcp.Server {
		return mcpServer
	}, nil)
	return originMiddleware(base)
}

func originMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		reqOrigin := r.Header.Get("Origin")
		if reqOrigin != "" && !origin.Allow(reqOrigin) {
			http.Error(w, "origin not allowed", http.StatusForbidden)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func registerTools(server *mcp.Server, s *Server) {
	mcp.AddTool(server, &mcp.Tool{Name: "simulate", Description: "Run reset + simulation"}, s.toolSimulate)
	mcp.AddTool(server, &mcp.Tool{Name: "step", Description: "Advance one clock cycle"}, s.toolStep)
	mcp.AddTool(server, &mcp.Tool{Name: "reset", Description: "Reset the simulator"}, s.toolReset)
	mcp.AddTool(server, &mcp.Tool{Name: "list_files", Description: "List workspace files"}, s.toolListFiles)
	mcp.AddTool(server, &mcp.Tool{Name: "read_file", Description: "Read a workspace file"}, s.toolReadFile)
	mcp.AddTool(server, &mcp.Tool{Name: "write_file", Description: "Write a workspace file"}, s.toolWriteFile)
	mcp.AddTool(server, &mcp.Tool{Name: "gate_count", Description: "Run Yosys stat gate count"}, s.toolGateCount)
	mcp.AddTool(server, &mcp.Tool{Name: "export_fpga", Description: "Export iCE40 bitstream"}, s.toolExportFPGA)
	mcp.AddTool(server, &mcp.Tool{Name: "get_uart", Description: "Get UART log text"}, s.toolGetUART)
	mcp.AddTool(server, &mcp.Tool{Name: "get_leds", Description: "Get LED state"}, s.toolGetLEDs)
	mcp.AddTool(server, &mcp.Tool{Name: "get_servos", Description: "Get PWM duty values"}, s.toolGetServos)
	mcp.AddTool(server, &mcp.Tool{Name: "get_framebuffer", Description: "Get framebuffer as base64"}, s.toolGetFramebuffer)
	mcp.AddTool(server, &mcp.Tool{Name: "get_waves", Description: "Get waveform samples"}, s.toolGetWaves)
	mcp.AddTool(server, &mcp.Tool{Name: "get_status", Description: "Get bench status"}, s.toolGetStatus)
	mcp.AddTool(server, &mcp.Tool{Name: "load_template", Description: "Load hello-gpu template"}, s.toolLoadTemplate)
}

func textResult(v any) (*mcp.CallToolResult, map[string]any, error) {
	b, err := json.Marshal(v)
	if err != nil {
		return nil, nil, err
	}
	return &mcp.CallToolResult{Content: []mcp.Content{&mcp.TextContent{Text: string(b)}}}, nil, nil
}

func (s *Server) toolSimulate(ctx context.Context, _ *mcp.CallToolRequest, _ struct{}) (*mcp.CallToolResult, any, error) {
	res := s.bench.Simulate(ctx)
	s.bench.RecordMCPTool("simulate", res.Pass)
	return textResult(res)
}

func (s *Server) toolStep(ctx context.Context, _ *mcp.CallToolRequest, _ struct{}) (*mcp.CallToolResult, any, error) {
	res := s.bench.Step(ctx)
	s.bench.RecordMCPTool("step", res.Pass)
	return textResult(res)
}

func (s *Server) toolReset(ctx context.Context, _ *mcp.CallToolRequest, _ struct{}) (*mcp.CallToolResult, any, error) {
	res := s.bench.Reset(ctx)
	s.bench.RecordMCPTool("reset", res.Pass)
	return textResult(res)
}

type pathArgs struct {
	Path string `json:"path" jsonschema:"relative workspace path"`
}

func (s *Server) toolListFiles(ctx context.Context, _ *mcp.CallToolRequest, _ struct{}) (*mcp.CallToolResult, any, error) {
	files, err := s.bench.Workspace().List()
	if err != nil {
		return nil, nil, err
	}
	s.bench.RecordMCPTool("list_files", nil)
	return textResult(map[string]any{"files": files})
}

func (s *Server) toolReadFile(ctx context.Context, _ *mcp.CallToolRequest, args pathArgs) (*mcp.CallToolResult, any, error) {
	content, err := s.bench.Workspace().Read(args.Path)
	if err != nil {
		return nil, nil, err
	}
	s.bench.RecordMCPTool("read_file", nil)
	return textResult(map[string]string{"path": args.Path, "content": content})
}

type writeArgs struct {
	Path    string `json:"path"`
	Content string `json:"content"`
}

func (s *Server) toolWriteFile(ctx context.Context, _ *mcp.CallToolRequest, args writeArgs) (*mcp.CallToolResult, any, error) {
	if err := s.bench.Workspace().Write(args.Path, args.Content); err != nil {
		return nil, nil, err
	}
	s.bench.RecordMCPTool("write_file", nil)
	return textResult(map[string]bool{"ok": true})
}

func (s *Server) toolGateCount(ctx context.Context, _ *mcp.CallToolRequest, _ struct{}) (*mcp.CallToolResult, any, error) {
	res := s.bench.GateCount(ctx)
	s.bench.RecordMCPTool("gate_count", nil)
	return textResult(res)
}

func (s *Server) toolExportFPGA(ctx context.Context, _ *mcp.CallToolRequest, _ struct{}) (*mcp.CallToolResult, any, error) {
	data, log, err := s.bench.ExportFPGA(ctx)
	if err != nil {
		return textResult(map[string]string{"error": err.Error(), "log": log})
	}
	s.bench.RecordMCPTool("export_fpga", nil)
	return textResult(map[string]string{
		"ok":       "true",
		"bytes":    fmt.Sprintf("%d", len(data)),
		"base64":   base64.StdEncoding.EncodeToString(data),
		"log_tail": tail(log, 4000),
	})
}

func (s *Server) toolGetUART(ctx context.Context, _ *mcp.CallToolRequest, _ struct{}) (*mcp.CallToolResult, any, error) {
	s.bench.RecordMCPTool("get_uart", nil)
	return textResult(map[string]string{"text": s.bench.UART()})
}

func (s *Server) toolGetLEDs(ctx context.Context, _ *mcp.CallToolRequest, _ struct{}) (*mcp.CallToolResult, any, error) {
	s.bench.RecordMCPTool("get_leds", nil)
	return textResult(map[string]int{"leds": s.bench.LEDs()})
}

func (s *Server) toolGetServos(ctx context.Context, _ *mcp.CallToolRequest, _ struct{}) (*mcp.CallToolResult, any, error) {
	s.bench.RecordMCPTool("get_servos", nil)
	return textResult(map[string][4]int{"duty": s.bench.Servos()})
}

func (s *Server) toolGetFramebuffer(ctx context.Context, _ *mcp.CallToolRequest, _ struct{}) (*mcp.CallToolResult, any, error) {
	s.bench.RecordMCPTool("get_framebuffer", nil)
	return textResult(map[string]string{"base64": base64.StdEncoding.EncodeToString(s.bench.Framebuffer())})
}

func (s *Server) toolGetWaves(ctx context.Context, _ *mcp.CallToolRequest, _ struct{}) (*mcp.CallToolResult, any, error) {
	s.bench.RecordMCPTool("get_waves", nil)
	return textResult(s.bench.Waves())
}

func (s *Server) toolGetStatus(ctx context.Context, _ *mcp.CallToolRequest, _ struct{}) (*mcp.CallToolResult, any, error) {
	s.bench.RecordMCPTool("get_status", nil)
	return textResult(s.bench.Status())
}

type templateArgs struct {
	Template string `json:"template" jsonschema:"template name, default hello-gpu"`
}

func (s *Server) toolLoadTemplate(ctx context.Context, _ *mcp.CallToolRequest, args templateArgs) (*mcp.CallToolResult, any, error) {
	name := args.Template
	if name == "" {
		name = "hello-gpu"
	}
	if err := s.bench.LoadTemplate(name); err != nil {
		return nil, nil, err
	}
	s.bench.RecordMCPTool("load_template", nil)
	return textResult(map[string]bool{"ok": true})
}

func tail(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[len(s)-n:]
}
