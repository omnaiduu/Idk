---
name: tiny-gpu-bench
description: Locked contracts, address map, PASS criteria, and QA checklist for Tiny GPU Bench. Use when editing HDL, firmware, Go API, SPA, Docker, or docs — not when only running the lab via MCP.
---

# Tiny GPU Bench agent skill (maintainers)

For **using** the running lab from Cursor (Lab2 MCP, custom `main.c`, simulate), load [tiny-gpu-lab](../tiny-gpu-lab/SKILL.md) instead.

This file is for **changing the product**.

## Quick facts

- **Run:** `docker compose up --build` → [http://127.0.0.1:8741](http://127.0.0.1:8741)
- **PASS source:** `templates/hello-gpu/expected.json` — never fake PASS
- **Sim script:** `./scripts/dev-sim.sh native` must exit 0 before claiming sim works
- **MCP:** `http://127.0.0.1:8741/mcp` — no auth, localhost only

## Address map (do not change)

| Device | Address |
|--------|---------|
| RAM | `0x0000_0000` |
| LEDs | `0x1000_0000` |
| BTN | `0x1000_0004` |
| UART TX | `0x2000_0000` |
| UART status | `0x2000_0004` bit0 |
| PWM n | `0x3000_0000 + 4*n` |
| GPU CMD | `0x4000_0000` |
| GPU STAT | `0x4000_0004` bit0 busy |

## UI colors

- Void `#14120b`, cream `#efece6`, ember `#ff6a2a`
- Dark only — no light mode

## Must not

- Tauri / desktop shell
- Publish `0.0.0.0:8741`
- MCP passwords
- Mock sim PASS
- `/lab` on omnaidu.com

## Docs to update when changing behavior

`README.md`, `docs/00-what.md` through `docs/05-mcp.md`, `docs/qa-log.md`, `docs/gaps.md`

See [PLAN.md](../PLAN.md) for the full master plan.
