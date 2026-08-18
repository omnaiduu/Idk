# Install

## Requirements

- **Docker Desktop** or Docker Engine with Compose v2
- **Git**
- About **4 GB RAM** and **2 CPUs** for the container (locked in `docker-compose.yml`)
- A browser on the same machine as Docker

You do **not** need Node, Go, or Verilator on the host — they live inside the image.

## Steps

```bash
git clone https://github.com/omnaiduu/tiny-gpu-bench.git
cd tiny-gpu-bench
docker compose up --build
```

Wait for the log line `tiny-gpu-bench listening on 0.0.0.0:8741`. The **first build** can take several minutes: the image installs apt packages (Verilator, Yosys, RISC-V gcc, nextpnr-ice40) and compiles the SPA and Go binary.

Open **[http://127.0.0.1:8741](http://127.0.0.1:8741)**.

## Port binding

Compose publishes **`127.0.0.1:8741:8741`** only. The Go server listens on `0.0.0.0` inside the container, but the host never exposes the port on all interfaces. Do not change this to `0.0.0.0:8741` — MCP has no password.

## Workspace volume

User edits and build artifacts persist in the Docker volume `bench-workspace` at `/data/workspace` inside the container.

## Health check

```bash
curl -s http://127.0.0.1:8741/health
# {"ok":true}
```

## Troubleshooting

| Symptom | Fix |
|--------|-----|
| Port in use | Stop other services on 8741 or change compose (not recommended for MCP docs). |
| OOM during first Run | Ensure Docker has ≥4 GB memory limit for the bench service. |
| Slow first Run | Verilator compiles the SoC on first simulate; later runs reuse `build/sim`. |
| Blank page | Wait for container healthy; hard-refresh after build completes. |

## Native dev (optional)

Developers with Verilator and `riscv64-unknown-elf-gcc` installed can run `./scripts/dev-sim.sh native` from the repo root for a CI-style PASS check without the full stack.
