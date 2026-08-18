package sim

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"

	"tiny-gpu-bench/internal/workspace"
)

// defaultRunCycles is the live-sim cap. PLAN.md mentioned 2M; hello-gpu firmware
// halt is ~2.42M so we allow 5M and stop early on ebreak.
const defaultRunCycles = 5_000_000

const fpgaTimeout = 5 * time.Minute

// Waves holds waveform samples from the last simulation.
type Waves struct {
	Clk     []int `json:"clk"`
	GPUBusy []int `json:"gpu_busy"`
	Led0    []int `json:"led0"`
	Pwm0    []int `json:"pwm0"`
}

// State is the cached bench output state.
type State struct {
	Cycles      int64
	Pass        *bool
	Outcome     string
	Running     bool
	Message     string
	UART        string
	LEDs        int
	Servos      [4]int
	Waves       Waves
	Framebuffer []byte
}

// Result is returned from simulate/step/reset operations.
type Result struct {
	Cycles  int64
	Pass    *bool
	Outcome string
	Message string
	Log     string
}

// Runner manages Verilator build/sim child processes.
type Runner struct {
	repoRoot  string
	scripts   string
	ws        *workspace.Workspace
	timeout   time.Duration
	templates string

	mu        sync.Mutex
	startMu   sync.Mutex
	cmd       *exec.Cmd
	stdin     io.WriteCloser
	snapCh    chan snapshot
	state     State
	lastStamp string
}

// NewRunner creates a simulation runner.
func NewRunner(repoRoot, scriptsDir string, ws *workspace.Workspace, templatesDir string, timeout time.Duration) *Runner {
	return &Runner{
		repoRoot:  repoRoot,
		scripts:   scriptsDir,
		ws:        ws,
		timeout:   timeout,
		templates: templatesDir,
	}
}

func (r *Runner) State() State {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.state
}

func (r *Runner) setRunning(v bool) {
	r.mu.Lock()
	r.state.Running = v
	if v {
		r.state.Pass = nil
		r.state.Outcome = ""
		r.state.Message = "running…"
	}
	r.mu.Unlock()
}

func (r *Runner) updateState(st State) {
	r.mu.Lock()
	r.state = st
	r.mu.Unlock()
}

// Simulate builds if needed, resets, and runs up to defaultRunCycles.
func (r *Runner) Simulate(ctx context.Context) (*Result, error) {
	if err := r.ws.EnsureDirs(); err != nil {
		return &Result{}, err
	}
	r.setRunning(true)
	defer r.setRunning(false)

	r.stopSim()
	logBuf, err := r.build(ctx)
	if err != nil {
		r.failState(err.Error())
		return &Result{Log: logBuf}, err
	}
	if err := r.ensureSim(ctx); err != nil {
		r.failState(err.Error())
		return &Result{Log: logBuf + err.Error()}, err
	}
	if _, err := r.sendCommand(ctx, map[string]any{"cmd": "reset"}); err != nil {
		r.failState(err.Error())
		return &Result{Log: logBuf}, err
	}
	snap, err := r.sendCommand(ctx, map[string]any{"cmd": "run", "cycles": defaultRunCycles})
	if err != nil {
		r.failState(err.Error())
		return &Result{Log: logBuf}, err
	}
	if err := r.loadDumps(); err != nil {
		r.failState(err.Error())
		return &Result{Log: logBuf}, err
	}
	pass, outcome, msg, perr := r.checkPass(snap.Halted)
	st := r.State()
	st.Cycles = snap.Cycles
	st.Pass = pass
	st.Outcome = outcome
	st.Message = msg
	r.updateState(st)
	res := &Result{Cycles: snap.Cycles, Pass: pass, Outcome: outcome, Message: msg, Log: logBuf}
	if perr != nil {
		res.Message = perr.Error()
	}
	return res, nil
}

func (r *Runner) failState(msg string) {
	st := r.State()
	pass := false
	st.Pass = &pass
	st.Outcome = OutcomeFail
	st.Message = msg
	st.Running = false
	r.updateState(st)
}

// Step executes a single clock cycle.
func (r *Runner) Step(ctx context.Context) (*Result, error) {
	if err := r.rebuildIfDirty(ctx); err != nil {
		return &Result{}, err
	}
	if err := r.ensureSim(ctx); err != nil {
		return &Result{}, err
	}
	snap, err := r.sendCommand(ctx, map[string]any{"cmd": "step"})
	if err != nil {
		return &Result{}, err
	}
	_ = r.loadDumps()
	st := r.State()
	st.Cycles = snap.Cycles
	r.updateState(st)
	return &Result{Cycles: snap.Cycles}, nil
}

// Reset resets the simulator state.
func (r *Runner) Reset(ctx context.Context) (*Result, error) {
	if err := r.rebuildIfDirty(ctx); err != nil {
		return &Result{}, err
	}
	if err := r.ensureSim(ctx); err != nil {
		return &Result{}, err
	}
	snap, err := r.sendCommand(ctx, map[string]any{"cmd": "reset"})
	if err != nil {
		return &Result{}, err
	}
	_ = r.loadDumps()
	st := r.State()
	st.Cycles = snap.Cycles
	st.Pass = nil
	st.Outcome = ""
	st.Message = ""
	st.UART = ""
	st.LEDs = 0
	st.Servos = [4]int{}
	st.Waves = Waves{}
	st.Framebuffer = nil
	r.updateState(st)
	return &Result{Cycles: snap.Cycles}, nil
}

// SetButton drives the virtual button input.
func (r *Runner) SetButton(down bool) error {
	r.mu.Lock()
	running := r.cmd != nil && r.stdin != nil
	r.mu.Unlock()
	if !running {
		return fmt.Errorf("simulator not running — press Run first")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	val := 0
	if down {
		val = 1
	}
	_, err := r.sendCommand(ctx, map[string]any{"cmd": "set_btn", "down": val})
	return err
}

type snapshot struct {
	OK     bool   `json:"ok"`
	Cycles int64  `json:"cycles"`
	Cycle  int64  `json:"cycle"`
	Halted bool   `json:"halted"`
	Error  string `json:"error"`
}

func (r *Runner) hdlDir() string {
	ws := r.ws.HDLDir()
	if _, err := os.Stat(filepath.Join(ws, "tiny_gpu_top.v")); err == nil {
		return ws
	}
	return filepath.Join(r.repoRoot, "hdl")
}

func (r *Runner) sourceStamp() string {
	var b strings.Builder
	paths := []string{filepath.Join(r.ws.FirmwareDir(), "main.c")}
	if files, err := filepath.Glob(filepath.Join(r.hdlDir(), "*.v")); err == nil {
		paths = append(paths, files...)
	}
	for _, p := range paths {
		info, err := os.Stat(p)
		if err != nil {
			continue
		}
		fmt.Fprintf(&b, "%s:%d:%d\n", p, info.Size(), info.ModTime().UnixNano())
	}
	return b.String()
}

func (r *Runner) rebuildIfDirty(ctx context.Context) error {
	if r.sourceStamp() == r.lastStamp && r.lastStamp != "" {
		if _, err := os.Stat(filepath.Join(r.ws.BuildDir(), "sim")); err == nil {
			return nil
		}
	}
	r.stopSim()
	_, err := r.build(ctx)
	return err
}

func (r *Runner) build(ctx context.Context) (string, error) {
	script := filepath.Join(r.scripts, "dev-sim.sh")
	cmd := exec.CommandContext(ctx, "bash", script, "build", r.ws.Root())
	cmd.Dir = r.repoRoot
	cmd.Env = append(os.Environ(),
		"REPO_ROOT="+r.repoRoot,
		"HDL_DIR="+r.hdlDir(),
	)
	var buf bytes.Buffer
	cmd.Stdout = &buf
	cmd.Stderr = &buf
	if err := runWithPG(ctx, cmd, r.timeout); err != nil {
		return buf.String(), fmt.Errorf("build failed: %w\n%s", err, buf.String())
	}
	r.lastStamp = r.sourceStamp()
	return buf.String(), nil
}

func (r *Runner) stopSim() {
	r.mu.Lock()
	cmd := r.cmd
	stdin := r.stdin
	r.cmd = nil
	r.stdin = nil
	r.snapCh = nil
	r.mu.Unlock()

	if stdin != nil {
		_, _ = stdin.Write([]byte("{\"cmd\":\"quit\"}\n"))
		_ = stdin.Close()
	}
	if cmd == nil || cmd.Process == nil {
		return
	}
	done := make(chan struct{})
	go func() {
		_ = cmd.Wait()
		close(done)
	}()
	select {
	case <-done:
		return
	case <-time.After(400 * time.Millisecond):
		killPG(cmd)
		<-done
	}
}

func (r *Runner) ensureSim(ctx context.Context) error {
	r.startMu.Lock()
	defer r.startMu.Unlock()

	r.mu.Lock()
	alive := r.cmd != nil && r.cmd.Process != nil
	r.mu.Unlock()
	if alive {
		return nil
	}

	simBin := filepath.Join(r.ws.BuildDir(), "sim")
	fwHex := filepath.Join(r.ws.BuildDir(), "firmware.hex")
	if _, err := os.Stat(simBin); err != nil {
		if _, err2 := r.build(ctx); err2 != nil {
			return err2
		}
	}

	cmd := exec.Command(simBin, "+firmware="+fwHex)
	cmd.Dir = r.ws.Root()
	cmd.Env = append(os.Environ(),
		"DUMP_DIR="+r.ws.DumpsDir(),
		"FIRMWARE_HEX="+fwHex,
	)
	setpg(cmd)

	stdin, err := cmd.StdinPipe()
	if err != nil {
		return err
	}
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return err
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		return err
	}
	if err := cmd.Start(); err != nil {
		return err
	}
	go io.Copy(io.Discard, stderr)

	r.mu.Lock()
	if r.cmd != nil && r.cmd.Process != nil {
		old := r.cmd
		r.mu.Unlock()
		killPG(old)
		_ = old.Wait()
		r.mu.Lock()
	}
	r.cmd = cmd
	r.stdin = stdin
	r.snapCh = make(chan snapshot, 8)
	r.mu.Unlock()

	go r.readSnapshots(stdout)
	_, _ = r.waitSnapshot(ctx, 5*time.Second)
	return nil
}

func (r *Runner) waitSnapshot(ctx context.Context, extra time.Duration) (snapshot, error) {
	r.mu.Lock()
	ch := r.snapCh
	r.mu.Unlock()
	if ch == nil {
		return snapshot{}, fmt.Errorf("simulator not running")
	}
	timer := time.NewTimer(extra)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return snapshot{}, ctx.Err()
	case snap := <-ch:
		return snap, nil
	case <-timer.C:
		return snapshot{}, fmt.Errorf("simulator response timeout")
	}
}

func (r *Runner) readSnapshots(rd io.Reader) {
	sc := bufio.NewScanner(rd)
	buf := make([]byte, 0, 64*1024)
	sc.Buffer(buf, 1024*1024)
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if line == "" {
			continue
		}
		var snap snapshot
		if json.Unmarshal([]byte(line), &snap) != nil {
			continue
		}
		cycles := snap.Cycles
		if cycles == 0 {
			cycles = snap.Cycle
		}
		st := r.State()
		st.Cycles = cycles
		r.updateState(st)
		r.mu.Lock()
		ch := r.snapCh
		r.mu.Unlock()
		if ch != nil {
			s := snapshot{OK: true, Cycles: cycles, Cycle: cycles, Halted: snap.Halted}
			select {
			case ch <- s:
			default:
			}
		}
	}
}

func (r *Runner) sendCommand(ctx context.Context, cmd map[string]any) (snapshot, error) {
	r.mu.Lock()
	stdin := r.stdin
	proc := r.cmd
	r.mu.Unlock()
	if stdin == nil || proc == nil {
		return snapshot{}, fmt.Errorf("simulator not running")
	}
	b, _ := json.Marshal(cmd)
	line := string(b) + "\n"
	done := make(chan error, 1)
	go func() {
		_, err := stdin.Write([]byte(line))
		done <- err
	}()
	select {
	case <-ctx.Done():
		r.stopSim()
		return snapshot{}, ctx.Err()
	case err := <-done:
		if err != nil {
			r.stopSim()
			return snapshot{}, err
		}
	}
	wait := 10 * time.Second
	if cmdName, _ := cmd["cmd"].(string); cmdName == "run" {
		wait = r.timeout
	}
	snap, err := r.waitSnapshot(ctx, wait)
	if err != nil {
		r.stopSim()
		return snapshot{}, err
	}
	return snap, nil
}

func (r *Runner) loadDumps() error {
	dumps := r.ws.DumpsDir()
	st := r.State()

	if b, err := os.ReadFile(filepath.Join(dumps, "uart.log")); err == nil {
		st.UART = string(b)
	}
	if b, err := os.ReadFile(filepath.Join(dumps, "leds.json")); err == nil {
		var v struct {
			LEDs int `json:"leds"`
		}
		if json.Unmarshal(b, &v) == nil {
			st.LEDs = v.LEDs
		}
	}
	if b, err := os.ReadFile(filepath.Join(dumps, "pwm.json")); err == nil {
		var v struct {
			Duty [4]int `json:"duty"`
		}
		if json.Unmarshal(b, &v) == nil {
			st.Servos = v.Duty
		}
	}
	if b, err := os.ReadFile(filepath.Join(dumps, "waves.json")); err == nil {
		_ = json.Unmarshal(b, &st.Waves)
	}
	if b, err := os.ReadFile(filepath.Join(dumps, "fb.bin")); err == nil {
		st.Framebuffer = b
	}
	r.updateState(st)
	return nil
}

func (r *Runner) checkPass(halted bool) (*bool, string, string, error) {
	expPath := filepath.Join(r.templates, "hello-gpu", "expected.json")
	b, err := os.ReadFile(expPath)
	if err != nil {
		pass := false
		return &pass, OutcomeFail, "missing expected.json", err
	}
	var exp Expected
	if err := json.Unmarshal(b, &exp); err != nil {
		pass := false
		return &pass, OutcomeFail, "invalid expected.json", err
	}
	st := r.State()
	if uartBytes, err := os.ReadFile(filepath.Join(r.ws.DumpsDir(), "uart.log")); err == nil {
		st.UART = string(uartBytes)
	}
	if ledsBytes, err := os.ReadFile(filepath.Join(r.ws.DumpsDir(), "leds.json")); err == nil {
		var v struct {
			LEDs int `json:"leds"`
		}
		if json.Unmarshal(ledsBytes, &v) == nil {
			st.LEDs = v.LEDs
		}
	}
	if pwmBytes, err := os.ReadFile(filepath.Join(r.ws.DumpsDir(), "pwm.json")); err == nil {
		var v struct {
			Duty [4]int `json:"duty"`
		}
		if json.Unmarshal(pwmBytes, &v) == nil {
			st.Servos = v.Duty
		}
	}
	if fb, err := os.ReadFile(filepath.Join(r.ws.DumpsDir(), "fb.bin")); err == nil {
		st.Framebuffer = fb
	}
	ok, outcome, msg := ScoreRun(exp, st.UART, st.LEDs, st.Servos, st.Framebuffer, halted)
	return ok, outcome, msg, nil
}

var statCells = regexp.MustCompile(`Number of cells:\s+(\d+)`)
var statWires = regexp.MustCompile(`Number of wire bits:\s+(\d+)`)

func (r *Runner) hdlFiles() ([]string, error) {
	hdl := r.hdlDir()
	files, err := filepath.Glob(filepath.Join(hdl, "*.v"))
	if err != nil {
		return nil, err
	}
	if len(files) == 0 {
		return nil, fmt.Errorf("no verilog files in %s", hdl)
	}
	return files, nil
}

// GateCount runs yosys stat on workspace HDL.
func (r *Runner) GateCount(ctx context.Context) (cells, wires int, msg string, err error) {
	files, err := r.hdlFiles()
	if err != nil {
		return 0, 0, "", err
	}
	var reads []string
	for _, f := range files {
		reads = append(reads, "read_verilog "+strconv.Quote(f))
	}
	script := strings.Join(reads, "; ") + "; hierarchy -check -top tiny_gpu_top; stat"
	cmd := exec.CommandContext(ctx, "yosys", "-p", script)
	var buf bytes.Buffer
	cmd.Stdout = &buf
	cmd.Stderr = &buf
	if err := runWithPG(ctx, cmd, r.timeout); err != nil {
		return 0, 0, buf.String(), fmt.Errorf("yosys failed: %w", err)
	}
	out := buf.String()
	if m := statCells.FindStringSubmatch(out); len(m) == 2 {
		cells, _ = strconv.Atoi(m[1])
	}
	if m := statWires.FindStringSubmatch(out); len(m) == 2 {
		wires, _ = strconv.Atoi(m[1])
	}
	if cells == 0 && wires == 0 {
		return 0, 0, out, fmt.Errorf("yosys produced no cell/wire counts — is yosys installed?")
	}
	return cells, wires, out, nil
}

// ExportFPGA synthesizes for iCE40 hx8k and returns a bitstream.
func (r *Runner) ExportFPGA(ctx context.Context) ([]byte, string, error) {
	build := r.ws.BuildDir()
	if err := os.MkdirAll(build, 0o755); err != nil {
		return nil, "", err
	}
	files, err := r.hdlFiles()
	if err != nil {
		return nil, "", err
	}
	jsonOut := filepath.Join(build, "fpga.json")
	ascOut := filepath.Join(build, "fpga.asc")
	binOut := filepath.Join(build, "fpga.bin")

	var reads []string
	for _, f := range files {
		reads = append(reads, "read_verilog "+strconv.Quote(f))
	}
	yosysScript := strings.Join(reads, "; ") +
		"; hierarchy -check -top tiny_gpu_top; synth_ice40 -top tiny_gpu_top; write_json " + strconv.Quote(jsonOut)

	cmd := exec.CommandContext(ctx, "yosys", "-p", yosysScript)
	var ybuf bytes.Buffer
	cmd.Stdout = &ybuf
	cmd.Stderr = &ybuf
	if err := runWithPG(ctx, cmd, fpgaTimeout); err != nil {
		return nil, ybuf.String(), fmt.Errorf("yosys synth failed: %w (need yosys in PATH / Docker image)", err)
	}

	ncmd := exec.CommandContext(ctx, "nextpnr-ice40", "--hx8k", "--package", "ct256", "--json", jsonOut, "--asc", ascOut)
	var nbuf bytes.Buffer
	ncmd.Stdout = &nbuf
	ncmd.Stderr = &nbuf
	if err := runWithPG(ctx, ncmd, fpgaTimeout); err != nil {
		return nil, ybuf.String() + nbuf.String(), fmt.Errorf("nextpnr failed: %w (need nextpnr-ice40; hx8k/ct256, no PCF)", err)
	}

	icmd := exec.CommandContext(ctx, "icepack", ascOut, binOut)
	var ibuf bytes.Buffer
	icmd.Stdout = &ibuf
	icmd.Stderr = &ibuf
	if err := runWithPG(ctx, icmd, r.timeout); err != nil {
		return nil, ybuf.String() + nbuf.String() + ibuf.String(), fmt.Errorf("icepack failed: %w", err)
	}
	data, err := os.ReadFile(binOut)
	if err != nil {
		return nil, ybuf.String() + nbuf.String(), err
	}
	return data, ybuf.String() + nbuf.String(), nil
}

func setpg(cmd *exec.Cmd) {
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
}

func runWithPG(ctx context.Context, cmd *exec.Cmd, timeout time.Duration) error {
	setpg(cmd)
	if err := cmd.Start(); err != nil {
		return err
	}
	done := make(chan error, 1)
	go func() { done <- cmd.Wait() }()
	t := timeout
	if deadline, ok := ctx.Deadline(); ok {
		if d := time.Until(deadline); d < t {
			t = d
		}
	}
	timer := time.NewTimer(t)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		killPG(cmd)
		<-done
		return ctx.Err()
	case err := <-done:
		return err
	case <-timer.C:
		killPG(cmd)
		<-done
		return fmt.Errorf("timeout after %s", t)
	}
}

func killPG(cmd *exec.Cmd) {
	if cmd.Process == nil {
		return
	}
	_ = syscall.Kill(-cmd.Process.Pid, syscall.SIGKILL)
}
