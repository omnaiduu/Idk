# Known gaps

Items unfinished or deferred. **No fake PASS** — if sim fails, the UI shows FAIL.

## Status (2026-08-17)

### Verified working

- Verilator + PicoRV32 + RISC-V C + DPI dumps — `./scripts/dev-sim.sh native` exits 0 with PASS
- Go API simulate/step/reset, MCP initialize, SPA dark UI, browser Run → PASS
- Concurrent Run returns HTTP 409

### Not verified in this environment

- **Docker compose end-to-end** — the cloud agent VM has no Docker daemon. `Dockerfile` and `docker-compose.yml` match PLAN.md; Om should run `docker compose up --build` locally.
- **Gate count / FPGA export in browser** — require yosys and nextpnr inside the container (not installed on the bare agent host). API returns clear errors when tools are missing.
- **Browser FAIL flow** (break `main.c`, Copy log) — not exercised in QA session; API PASS/FAIL logic is implemented via `expected.json`.

### Plan note: cycle limit

PLAN.md mentions a 2M default Run cap; hello-gpu firmware halts at ~**2.42M** cycles. The server runs up to **5M** cycles (same as `dev-sim.sh`) and stops early on `ebreak` halt. Consider shortening firmware if a strict 2M cap is required.

### Not implemented (by design)

- Real USB motor control — servos are drawn only
- Extra templates beyond `hello-gpu`
- Light mode / theme toggle
- MCP authentication or TLS

If a future agent stalls on Verilator flags or GPU busy never clearing, document the exact error here — do **not** ship `MOCK_SIM=1` as real PASS.
