# Tiny GPU Bench

Tiny GPU Bench is a browser lab for writing Verilog and RISC-V C, pressing Run, and seeing whether a tiny digital machine actually works: UART text, blinking LEDs, a 64×64 pixel screen, drawn servo arms, and a clear PASS or FAIL. One Docker container runs Verilator, Yosys, and the RISC-V toolchain; the React UI talks to a Go server over HTTP. This is an open write-and-test lab, not Cadence and not a phone factory.

## Install and run

You need Docker and Git.

```bash
git clone https://github.com/omnaiduu/tiny-gpu-bench.git
cd tiny-gpu-bench
docker compose up --build
```

Open [http://127.0.0.1:8741](http://127.0.0.1:8741). The first build is slow while Verilator compiles the SoC. Press **Run** on the hello-gpu template; you should see `hello from tiny-gpu` in the UART log, lit LEDs, an OM rectangle on the screen, and a green PASS pill.

## MCP (Cursor)

The same Go process serves MCP on localhost with no password. Add this to Cursor:

```json
{
  "mcpServers": {
    "tiny-gpu-bench": {
      "url": "http://127.0.0.1:8741/mcp"
    }
  }
}
```

See [docs/05-mcp.md](docs/05-mcp.md) for curl examples. Agents: read [`.cursor/skills/tiny-gpu-lab/SKILL.md`](.cursor/skills/tiny-gpu-lab/SKILL.md) for how to drive the bench (hello-gpu, custom C, PASS vs FAIL). **Never publish port 8741 to the public internet** — there is no authentication.

## What this is not

Not Synopsys, not TSMC PDKs, not NVIDIA CUDA, not a desktop installer, and not a replacement for professional EDA. It is a friendly sandbox for one PicoRV32 SoC with a tiny GPU painter block.

## Docs

Blog-ready material lives in [docs/](docs/):

- [00-what.md](docs/00-what.md) — overview
- [01-install.md](docs/01-install.md) — Docker setup
- [02-use.md](docs/02-use.md) — UI walkthrough
- [03-how-it-works.md](docs/03-how-it-works.md) — architecture
- [04-decisions.md](docs/04-decisions.md) — decision log
- [05-mcp.md](docs/05-mcp.md) — MCP tools and curl
- [`.cursor/skills/tiny-gpu-lab/SKILL.md`](.cursor/skills/tiny-gpu-lab/SKILL.md) — agent skill for running the lab

License: Apache-2.0 for our code. Verilator and other tools keep their own licenses inside the Docker image.
