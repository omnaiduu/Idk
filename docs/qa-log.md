# QA log

Browser and API testing log for Tiny GPU Bench implementation.

## 2026-08-18 — Lab2 MCP hello-gpu demo

### Lab2 MCP (`tiny-gpu-bench` over Streamable HTTP)

Used the live Lab2 MCP tools against the hello-gpu template (`firmware/main.c`: UART hello, LED chaser, OM glyphs, PWM 32/96/160/224).

| Check | Result | Notes |
|------|--------|-------|
| MCP `initialize` | PASS | protocol **2025-11-25**, server `tiny-gpu-bench` |
| `load_template` / workspace files | PASS | `firmware/main.c` is the hello-gpu demo |
| `simulate` | PASS | **2,423,702** cycles, `last_mcp_tool=simulate` |
| `get_uart` | PASS | `hello from tiny-gpu` |
| `get_leds` | PASS | **170** (`0xAA`) |
| `get_servos` | PASS | `[32, 96, 160, 224]` |
| `get_framebuffer` | PASS | 4096 bytes, sha256 matches `expected.json` |
| `get_waves` | PASS | 1024 samples each for clk / gpu_busy / led0 / pwm0 |

Cursor Cloud `CallMcpTool` failed on the wire shape `"structuredContent": null` (expected a JSON object). MCP tools still executed; results were read via MCP `tools/call` curl. Fixed handlers to always return a JSON object as structured output.

Custom firmware that **halts** is now **RAN**, not FAIL. FAIL is reserved for compile/timeout/CPU-never-halted. PASS remains the hello-gpu golden contract.

---

## 2026-08-18 — UI polish + backend hardening

### Environment

- Native Go server `http://127.0.0.1:8741` with host Verilator 5.020 + RISC-V gcc
- SPA production build (`apps/web/dist`)
- No Docker daemon; yosys/nextpnr not on host

### API

| Test | Result | Notes |
|------|--------|-------|
| `GET /health` | PASS | `{"ok":true}` |
| `POST /api/simulate` | PASS | `pass: true`, **2,423,702** cycles, UART hello, HDL_DIR=`workspace-data/hdl` |
| Concurrent simulate | PASS | HTTP **409** `busy` / other request PASS |
| MCP initialize (no auth headers) | PASS | protocol **2025-11-25**, server `tiny-gpu-bench` |
| `go test ./...` | PASS | origin, PASS evaluator, workspace jail, CORS, `toSimResult` |

### Browser checklist (PLAN.md)

| # | Check | Result | Notes |
|---|-------|--------|-------|
| 1 | Dark layout 1280+, no white flash | PASS | Void first paint; idle `—` pill (not leftover PASS) |
| 2 | 3-step overlay dismiss + localStorage | PASS | Corner card, not a blocking modal; Run stays clickable; stays gone on reload |
| 3 | Run → UART hello, LEDs, OM, PASS | PASS | ~2.42M cycles; 4 ember LEDs (170); OM on CRT; servos 32/96/160/224 |
| 4 | Break `main.c` → FAIL + Copy | PASS | Deleted semicolon; FAIL pill; slide-over log; Copy; no `alert` |
| 5 | Fix → PASS again | PASS | Undo + Run |
| 6 | GPU vs CPU Monaco switch | PASS | Also RAM/UART/GPIO/PWM files |
| 7 | How big? | PASS (error) | Clear yosys-missing dialog; no crash |
| 8 | Export FPGA | PASS (error) | Slide-over with log; title is “FPGA export failed” |
| 9 | prefers-reduced-motion | PASS | CSS + hook still in SPA |
| 10 | MCP curl | PASS | initialize 200, protocol 2025-11-25 |
| 11 | Second Run → 409 | PASS | API concurrent 409; UI disables Run while busy |

CRT is integer **×4 (256px)** and fits the 280px rail (was overflowing at ×8).

### Bugs found this pass (fixed in code)

1. CRT 512px overflow in 280px rail
2. Onboarding Dialog blocked Run (focus trap)
3. Stale PASS on first paint from `/api/status`
4. Monaco light flash before `onMount`
5. Error panel could not scroll (`min-h-0` missing)
6. 409 treated as FAIL in the pill
7. Workspace HDL ignored; Verilator cache skipped rebuilds
8. `toSimResult` nil panic; job lock holes; zombie sim processes
9. Reset forced FAIL; `halted` not checked; waves were first-1024 not last-1024
10. FPGA export errors labeled “Simulation error”
11. Gate-count busy briefly looked like a sim RUN on the pill

### Remaining for Om (Docker machine)

- Full `docker compose up --build` smoke test
- Gate count numbers and FPGA `.bin` download inside the container

---

## 2026-08-17 — Initial implementation QA

### Environment

- Native `./scripts/dev-sim.sh native` — **PASS** (UART hello, LEDs 170, PWM [32,96,160,224], fb_sha256 match, halted at 2,423,702 cycles)
- Docker compose — not run in cloud agent VM (Docker unavailable); Dockerfile and compose verified in repo
- Go server at `http://127.0.0.1:8741` with host Verilator + RISC-V gcc

### API tests

| Test | Result | Notes |
|------|--------|-------|
| `GET /health` | PASS | `{"ok":true}` |
| `POST /api/simulate` | PASS | PASS after raising default cycles to 5M (firmware halts ~2.42M) |
| `GET /api/uart` | PASS | Contains `hello from tiny-gpu` |
| `GET /api/leds` | PASS | `170` |
| Second simulate while busy | PASS | HTTP **409**, `error: busy` |
| `POST /api/gate-count` | N/A host | yosys not on host; works inside Docker image |
| `POST /api/export-fpga` | N/A host | Requires yosys/nextpnr in Docker |
| MCP initialize | PASS | Streamable HTTP, protocol 2025-11-25 |

### Browser checklist (PLAN.md)

| # | Check | Result | Notes |
|---|-------|--------|-------|
| 1 | Dark layout 1280+, no white flash | PASS | Void background on first paint |
| 2 | 3-step overlay dismiss + localStorage | PASS | All three steps shown; dismiss works |
| 3 | Run → UART hello, LEDs, OM rect, PASS | PASS | 2,423,702 cycles, green PASS pill |
| 4 | Break main.c → FAIL + Copy | not run | Server-side PASS logic verified via API |
| 5 | Fix → PASS again | not run | |
| 6 | GPU vs CPU Monaco switch | PASS | `firmware/main.c` ↔ `hdl/gpu.v` |
| 7 | How big? | N/A host | Needs Docker yosys |
| 8 | Export FPGA | N/A host | Needs Docker nextpnr |
| 9 | prefers-reduced-motion | PASS | Hook in SPA; animations respect media query |
| 10 | MCP curl | PASS | initialize returns server capabilities |
| 11 | Second Run → 409 | PASS | Fixed after `stopSim()` before rebuild |

Recording: `/workspace/qa_recording_20260817_132023.mp4`

### Bugs found and fixed

1. **Verilator concurrent build crash** — `attempted to destroy locked Thread Pool` when rebuilding while obj_dir hot. Fixed: cache Verilator binary with flock; `-j1` in Makefile.
2. **Sim binary "Text file busy"** — `cp` over running executable during rebuild. Fixed: `stopSim()` before build; atomic `cp` + `mv`.
3. **PASS fail at 2M cycles** — Firmware halts at ~2.42M cycles. Fixed: `defaultRunCycles = 5_000_000` (still stops early on halt).
4. **409 not returned on concurrent Run** — Race when sim still building. Fixed: job mutex + stop sim before rebuild.

### Remaining for Om (Docker machine)

- Full `docker compose up --build` smoke test
- Gate count and FPGA export inside container
- Break/fix `main.c` browser FAIL flow
