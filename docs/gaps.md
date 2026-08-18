# Known gaps

Items unfinished or deferred. **No fake PASS** — if sim fails, the UI shows FAIL.

## Status (2026-08-18 polish)

### Fixed in this pass

- Stale PASS pill on first paint (idle until this browser session runs, or until `/api/status` reports a live job)
- Onboarding no longer uses a modal Dialog that trapped focus and blocked **Run**
- CRT was 512px (`×8`) inside a 280px rail — now integer `×4` (256px) with a scanline/vignette
- Student HDL edits in `workspace/hdl` were ignored (Verilator always built repo `hdl/` and skipped rebuild if a binary existed)
- `toSimResult` nil panic on Step/Reset errors
- Step/Reset/LoadTemplate skipped the exclusive job lock (flaky 409)
- Sim process not `Wait`ed → zombies; timeout did not kill the live sim group
- Reset forced `pass: false` (FAIL after Reset)
- `checkPass` ignored `halted` from `expected.json`
- Waves kept the **first** 1024 samples; now a ring of the **last** 1024
- Native `HOST` default was `0.0.0.0`; now `127.0.0.1` (Docker compose still sets `0.0.0.0`)
- Workspace template load on every boot wiped HDL; now `LoadTemplateIfEmpty`
- SPA `filepath.Join` with cleaned URL paths could ignore `webRoot` on Unix
- Go unit tests for origin, PASS evaluator, workspace jail, CORS, `toSimResult`

### Verified working (prior + this pass)

- Verilator + PicoRV32 + RISC-V C + DPI dumps — `./scripts/dev-sim.sh native` exits 0 with PASS when tools are installed
- Go API simulate/step/reset, MCP initialize, SPA dark UI
- Concurrent Run returns HTTP 409
- `go test ./...` on the Go packages above

### Not verified in this environment

- **Docker compose end-to-end** — this cloud agent VM has no Docker daemon. `Dockerfile` and `docker-compose.yml` match PLAN.md; Om should run `docker compose up --build` locally.
- **Gate count / FPGA export in browser** — require yosys and nextpnr inside the container (not installed on the bare agent host). API returns clear errors when tools are missing. FPGA path now uses a 5-minute timeout (was capped at the 120s sim timeout).
- **Browser FAIL flow** (break `main.c`, Copy log) — exercise after a real Verilator Run.

### Plan note: cycle limit

PLAN.md mentions a 2M default Run cap; hello-gpu firmware halts at ~**2.42M** cycles. The server still runs up to **5M** cycles (same as `dev-sim.sh`) and stops early on `ebreak` halt. Shortening firmware would change the golden `fb_sha256`; left as-is.

### Not implemented (by design)

- Real USB motor control — servos are drawn only
- Extra templates beyond `hello-gpu`
- Light mode / theme toggle
- MCP authentication or TLS

If a future agent stalls on Verilator flags or GPU busy never clearing, document the exact error here — do **not** ship `MOCK_SIM=1` as real PASS.
