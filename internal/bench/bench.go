package bench

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"tiny-gpu-bench/internal/sim"
	"tiny-gpu-bench/internal/workspace"
)

// Status is returned by GET /api/status and MCP get_status.
type Status struct {
	OK          bool   `json:"ok"`
	Cycles      int64  `json:"cycles"`
	Pass        *bool  `json:"pass"`
	Outcome     string `json:"outcome,omitempty"`
	Running     bool   `json:"running"`
	Busy        bool   `json:"busy,omitempty"`
	Message     string `json:"message,omitempty"`
	LastMCPTool string `json:"last_mcp_tool,omitempty"`
	LastMCPPass *bool  `json:"last_mcp_pass,omitempty"`
}

// SimulateResult is returned by simulate/step/reset endpoints.
type SimulateResult struct {
	OK      bool   `json:"ok"`
	Pass    *bool  `json:"pass,omitempty"`
	Outcome string `json:"outcome,omitempty"`
	Cycles  int64  `json:"cycles,omitempty"`
	Message string `json:"message,omitempty"`
	Error   string `json:"error,omitempty"`
	Log     string `json:"log,omitempty"`
}

// GateCountResult is returned by gate-count.
type GateCountResult struct {
	OK      bool   `json:"ok"`
	Cells   int    `json:"cells,omitempty"`
	Wires   int    `json:"wires,omitempty"`
	Message string `json:"message,omitempty"`
	Error   string `json:"error,omitempty"`
}

// Bench coordinates simulation jobs and cached dump state.
type Bench struct {
	ws     *workspace.Workspace
	runner *sim.Runner

	mu          sync.Mutex
	jobRunning  bool
	lastMCPTool string
	lastMCPPass *bool
}

// New creates a bench service.
func New(ws *workspace.Workspace, runner *sim.Runner) *Bench {
	return &Bench{ws: ws, runner: runner}
}

func (b *Bench) Workspace() *workspace.Workspace { return b.ws }
func (b *Bench) Runner() *sim.Runner             { return b.runner }

func (b *Bench) beginJob() error {
	b.mu.Lock()
	defer b.mu.Unlock()
	if b.jobRunning {
		return fmt.Errorf("busy")
	}
	b.jobRunning = true
	return nil
}

func (b *Bench) endJob() {
	b.mu.Lock()
	b.jobRunning = false
	b.mu.Unlock()
}

func (b *Bench) IsBusy() bool {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.jobRunning
}

func (b *Bench) Status() Status {
	st := b.runner.State()
	b.mu.Lock()
	defer b.mu.Unlock()
	return Status{
		OK:          true,
		Cycles:      st.Cycles,
		Pass:        st.Pass,
		Outcome:     st.Outcome,
		Running:     st.Running,
		Busy:        b.jobRunning,
		Message:     st.Message,
		LastMCPTool: b.lastMCPTool,
		LastMCPPass: b.lastMCPPass,
	}
}

func (b *Bench) recordMCP(tool string, pass *bool) {
	b.mu.Lock()
	b.lastMCPTool = tool
	b.lastMCPPass = pass
	b.mu.Unlock()
}

func (b *Bench) toSimResult(r *sim.Result, err error) SimulateResult {
	if err != nil {
		if err.Error() == "busy" {
			return SimulateResult{OK: false, Error: "busy", Message: "another job is running"}
		}
		log := ""
		if r != nil {
			log = r.Log
		}
		return SimulateResult{OK: false, Error: err.Error(), Log: log}
	}
	if r == nil {
		return SimulateResult{OK: false, Error: "internal error"}
	}
	res := SimulateResult{OK: true, Cycles: r.Cycles, Outcome: r.Outcome, Message: r.Message, Log: r.Log}
	if r.Pass != nil {
		res.Pass = r.Pass
	}
	return res
}

// Simulate runs reset + up to 5M cycles (halts earlier on ebreak).
func (b *Bench) Simulate(ctx context.Context) SimulateResult {
	if err := b.beginJob(); err != nil {
		return SimulateResult{OK: false, Error: "busy", Message: "another job is running"}
	}
	defer b.endJob()
	r, err := b.runner.Simulate(ctx)
	return b.toSimResult(r, err)
}

// Step advances one clock cycle.
func (b *Bench) Step(ctx context.Context) SimulateResult {
	if err := b.beginJob(); err != nil {
		return SimulateResult{OK: false, Error: "busy", Message: "another job is running"}
	}
	defer b.endJob()
	r, err := b.runner.Step(ctx)
	return b.toSimResult(r, err)
}

// Reset resets the live simulator.
func (b *Bench) Reset(ctx context.Context) SimulateResult {
	if err := b.beginJob(); err != nil {
		return SimulateResult{OK: false, Error: "busy", Message: "another job is running"}
	}
	defer b.endJob()
	r, err := b.runner.Reset(ctx)
	return b.toSimResult(r, err)
}

// SetButton drives the virtual button. Does not take the job lock so it can
// be pressed during a live step session; Simulate holds the lock for the full run.
func (b *Bench) SetButton(down bool) error {
	if b.IsBusy() && b.runner.State().Running {
		// Mid-simulate stdin is owned by the run command; skip rather than race.
		return fmt.Errorf("busy — wait until Run finishes, then use Step to poke the button")
	}
	return b.runner.SetButton(down)
}

// GateCount runs yosys stat.
func (b *Bench) GateCount(ctx context.Context) GateCountResult {
	if err := b.beginJob(); err != nil {
		return GateCountResult{OK: false, Error: "busy", Message: "another job is running"}
	}
	defer b.endJob()
	cells, wires, msg, err := b.runner.GateCount(ctx)
	if err != nil {
		return GateCountResult{OK: false, Error: err.Error(), Message: msg}
	}
	return GateCountResult{OK: true, Cells: cells, Wires: wires, Message: msg}
}

// ExportFPGA synthesizes and packs a bitstream.
func (b *Bench) ExportFPGA(ctx context.Context) ([]byte, string, error) {
	if err := b.beginJob(); err != nil {
		return nil, "", fmt.Errorf("busy")
	}
	defer b.endJob()
	return b.runner.ExportFPGA(ctx)
}

// LoadTemplate loads a template into the workspace.
func (b *Bench) LoadTemplate(name string) error {
	if err := b.beginJob(); err != nil {
		return fmt.Errorf("busy")
	}
	defer b.endJob()
	return b.ws.LoadTemplate(name)
}

// Dump accessors
func (b *Bench) UART() string        { return b.runner.State().UART }
func (b *Bench) LEDs() int           { return b.runner.State().LEDs }
func (b *Bench) Servos() [4]int      { return b.runner.State().Servos }
func (b *Bench) Waves() sim.Waves    { return b.runner.State().Waves }
func (b *Bench) Framebuffer() []byte { return b.runner.State().Framebuffer }

// RecordMCPTool updates MCP drawer state.
func (b *Bench) RecordMCPTool(tool string, pass *bool) { b.recordMCP(tool, pass) }

// DefaultTimeout returns the configured simulation timeout.
func DefaultTimeout(sec int) time.Duration {
	if sec <= 0 {
		sec = 120
	}
	return time.Duration(sec) * time.Second
}

// WavesJSON marshals waves for API responses.
func WavesJSON(w sim.Waves) ([]byte, error) {
	return json.Marshal(w)
}
