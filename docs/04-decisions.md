# Decision log

Timeline of locked choices for Tiny GPU Bench. Agents should append dates and PR links as the project evolves.

1. **Lab not factory** — Open Verilator/Yosys flow, not Cadence/Synopsys paid shops. *(Aug 2026)*
2. **Tiny GPU = painter** — 64×64 RGB332 framebuffer with memory-mapped commands; not NVIDIA CUDA. *(Aug 2026)*
3. **First shape was Tauri desktop** — Dropped 17 Aug 2026; install footprint too large; Rust waiters existed only for Tauri shell. *(Aug 2026)*
4. **Client-server + Docker** — Student needs Docker + browser; chip tools stay in the image. *(Aug 2026)*
5. **Not Workers / Vercel Functions** — Cannot hold Verilator + clang in serverless limits. *(Aug 2026)*
6. **Not full OSS CAD Suite** — Only Verilator, clang, Yosys, nextpnr-ice40, icepack, RISC-V gcc from apt. *(Aug 2026)*
7. **Node/Hono replaced by Go** — Spawn/timeout/kill for child processes; official MCP Go SDK; one static binary in Docker. *(Om, 17 Aug 2026)*
8. **MCP no password** — Safe only because compose publishes `127.0.0.1:8741`; Origin check on API and MCP. *(Aug 2026)*
9. **SPA (Vite + React 19)** — One-page lab, not Next.js SSR; Go serves `apps/web/dist`. *(Aug 2026)*
10. **PASS is numbers** — `expected.json` with UART substring, LED/PWM values, `fb_sha256`, `halted`; never fake PASS. *(Aug 2026)*

11. **UI + backend polish (18 Aug 2026)** — CRT scale fits the 280px rail; onboarding is a corner card that never blocks Run; PASS pill stays idle until this session actually runs; HDL builds from the workspace copy; sim process groups are waited/killed on timeout; native Go bind defaults to `127.0.0.1` (Docker still sets `HOST=0.0.0.0`).

## Why Go?

Go’s `os/exec` with process groups, timeouts, and a single static binary fits the “waiter” role: start Verilator, kill on timeout, parse JSON lines. The official [`github.com/modelcontextprotocol/go-sdk`](https://github.com/modelcontextprotocol/go-sdk) provides Streamable HTTP MCP. Rust would also work; Go was chosen after dropping Tauri for faster shipping.

## Why Docker?

Verilator, Yosys, nextpnr, and RISC-V gcc are heavy native dependencies. One `docker compose up` gives every student the same kitchen without installing EDA tools on the host.

## Why no MCP auth?

Localhost-only binding is the security boundary. Adding tokens would complicate Cursor setup without meaningful gain when the port is not exposed publicly.

## Why this GPU?

Teaching needs visible pixels when C runs. A minimal 4-pixels-per-cycle painter with CLEAR / SET_COLOR / PUT_PIXEL / FILL_RECT is enough to draw rectangles and letters without implementing a shader ISA.

## Why SPA not omnaidu.com /lab?

The personal site stays a writing index. Bench code and docs live in this repo; blog posts come later from `docs/`.
