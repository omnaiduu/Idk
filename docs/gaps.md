# Known gaps

Items unfinished or deferred. **No fake PASS** — if sim fails, the UI shows FAIL.

## Status (2026-08-17)

### Resolved in this implementation

- Verilator + PicoRV32 + RISC-V C + DPI dumps — `./scripts/dev-sim.sh native` exits 0 with PASS.
- Go backend, React SPA, Dockerfile, and compose file created per PLAN.md.

### Pending verification

- **Full Docker compose browser QA** — container build and interactive checklist in `qa-log.md` must be completed on the target machine. First Run inside Docker may take 60–120s while Verilator compiles.
- **FPGA export timing** — `export_fpga` depends on Yosys/nextpnr in the runtime image; verify `.bin` download on a machine with sufficient disk and CPU.

### Not implemented (by design)

- Real USB motor control — servos are drawn only.
- Extra templates beyond `hello-gpu`.
- Light mode / theme toggle.
- MCP authentication or TLS.

If a future agent stalls on Verilator flags, PicoRV32 bus timing, or GPU busy never clearing, document the exact error, files touched, and attempts here — do **not** ship `MOCK_SIM=1` as real PASS.
