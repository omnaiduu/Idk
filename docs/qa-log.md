# QA log

Browser and API testing log for Tiny GPU Bench implementation.

## 2026-08-17 — Initial implementation QA

### Environment

- `./scripts/dev-sim.sh native` — **PASS** (UART hello, LEDs 170, PWM [32,96,160,224], fb_sha256 match, halted)
- Docker compose build — pending full container QA in this session

### Checklist (from PLAN.md)

| # | Check | Result | Notes |
|---|-------|--------|-------|
| 1 | Dark layout 1280/1440, no white flash | pending | Browser QA after compose up |
| 2 | 3-step overlay dismiss + localStorage | pending | |
| 3 | Run → UART hello, LEDs, OM rect, PASS | pending | Native sim PASS confirmed |
| 4 | Break main.c → FAIL + Copy, no alert | pending | |
| 5 | Fix → PASS again | pending | |
| 6 | GPU vs CPU Monaco file switch | pending | |
| 7 | How big? returns numbers | pending | Requires yosys in container |
| 8 | Export FPGA .bin or clear error | pending | |
| 9 | prefers-reduced-motion usable | pending | Hook implemented in SPA |
| 10 | MCP curl initialize + simulate | pending | |
| 11 | Second Run → 409 or disabled | pending | Go returns 409 when busy |

### Fixes applied during build

- **Merge conflict** in `scripts/dev-sim.sh` — unified Go `build`/`run`/`native` modes with HDL branch PASS checks (leds_final, pwm_duty, halted).
- **Merge conflict** in `expected.json` — kept golden hash from verified Verilator run (`d83705b8…`).

### Open items

See [gaps.md](gaps.md) for anything still unfinished after full Docker + browser pass.
