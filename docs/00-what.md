# What is Tiny GPU Bench?

Tiny GPU Bench is a browser-based digital lab. You edit Verilog (the machine) and C (the program running on a CPU inside that machine), press Run, and the server simulates the design with Verilator. The UI shows UART output, eight LEDs, four drawn servo arms, a 64×64 pixel screen, waveforms, and a verdict: **PASS** (hello-gpu golden), **RAN** (your firmware halted — custom code), or **FAIL** (it did not run).

## What it does

- **Simulate** — Verilator runs a PicoRV32 SoC with RAM, GPIO, UART, PWM, and a tiny GPU that paints pixels when C writes to memory-mapped registers.
- **How big?** — Yosys reports gate counts for the synthesized design.
- **Export FPGA** — Builds a `.bin` bitstream for iCE40 hx8k (no USB programming in the lab).
- **MCP** — Cursor can call the same Run, read, and write tools over HTTP on `http://127.0.0.1:8741/mcp` with no password (localhost only).

## What it is not

- Not Cadence, Synopsys, or Siemens EDA.
- Not NVIDIA CUDA — the “GPU” here is a simple pixel painter block, not a shader processor.
- Not analog simulation, OpenROAD, or a TSMC PDK flow.
- Not a desktop app (no Tauri, Electron, or installers).
- Not a cloud serverless sim — Verilator and gcc need a real Linux box inside Docker.
- Not safe on the public internet — MCP has no auth; compose binds `127.0.0.1:8741` only.

## One template

The repo ships **hello-gpu**: UART prints `hello from tiny-gpu`, LEDs chase, the GPU draws rectangles and the letters **OM**, and four PWM channels hold duties 32/96/160/224. PASS is checked against committed golden hashes, not vibes.
