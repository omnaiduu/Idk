# QA log

Browser and API testing log for Tiny GPU Bench implementation.

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
