---
name: tiny-gpu-lab
description: Operate Tiny GPU Bench (often the Lab2 or tiny-gpu-bench MCP) — write RISC-V C or Verilog, simulate, and read UART/LEDs/GPU. Use when the user asks to run the bench, hello-gpu, custom firmware, MCP simulate, Lab2, or a lab exercise on the fake chip.
---

# Tiny GPU Lab — agent skill

This is **how to use the running lab**, not how to rewrite the Go/React server.

The machine is a PicoRV32 SoC in Verilator. **C** (`firmware/main.c`) is the program on the CPU. **Verilog** (`hdl/*.v`) is the machine. They meet on a fixed address map. The “GPU” paints a 64×64 RGB332 framebuffer — not CUDA.

MCP is usually named **`Lab2`** or **`tiny-gpu-bench`**. Same tools. URL when local: `http://127.0.0.1:8741/mcp` (no auth, localhost only).

## When to load this

User wants to: run hello-gpu, write custom C, blink LEDs, print UART, draw pixels, step/reset, gate count, export FPGA, or “use Lab2 MCP”.

If they want to **change the bench product** (API, SPA, Docker, PASS evaluator), use `.cursor/skills/tiny-gpu-bench/SKILL.md` instead.

## First actions

1. `get_status` — if `busy`/`running`, wait; do not start a second `simulate` (HTTP 409).
2. `list_files` / `read_file` `firmware/main.c` — see what is loaded.
3. Then either restore the demo or write new firmware (below).

Do **not** claim PASS from vibes. Read the simulate JSON plus `get_uart`, `get_leds`, `get_servos`. If Cursor rejects a tool result (`structuredContent` null), the tool often still ran — confirm with `get_status` / `get_uart`.

## Standard workflows

### A. Smoke the demo (hello-gpu)

```
load_template { "template": "hello-gpu" }
simulate
get_uart / get_leds / get_servos / get_framebuffer
```

Expect **PASS** when:

- UART contains `hello from tiny-gpu`
- LEDs final **170** (`0xAA`)
- PWM **[32, 96, 160, 224]**
- CPU halted (~2.42M cycles with the stock chaser)
- Framebuffer SHA-256 matches `templates/hello-gpu/expected.json`

### B. Run custom C (usual request)

```
write_file { "path": "firmware/main.c", "content": "<full file>" }
simulate
get_uart
get_leds
get_servos
```

Keep `hdl/` alone unless the user asked to change the machine.

**PASS / RAN / FAIL:** `expected.json` is the **hello-gpu checklist**. Matching it is **PASS**. Custom firmware that **halts** is **RAN** (UART/LEDs/screen are yours). **FAIL** is compile error, timeout, or the CPU never halted — not “wrong UART string.”

Judge custom code by:

- `ok: true` and `outcome: "ran"` (or `"pass"` for hello-gpu)
- A halt (tens of thousands of cycles if you skip long `delay()`)
- UART / LEDs / PWM / framebuffer matching **what you wrote**
- Message `RAN: …` means the demo checklist differed — not a crash

Compile error / timeout / never-halted → **FAIL**. Golden mismatch after a clean halt → **RAN**.

When done on a **shared** bench, restore: `load_template { "template": "hello-gpu" }`.

### C. Change the machine (Verilog)

`write_file` paths like `hdl/gpu.v`. Next `simulate` rebuilds Verilator (slow). Do not invent a second address map.

## MCP tools

| Tool | Arguments | Notes |
|------|-----------|--------|
| `get_status` | — | Busy, cycles, pass, last MCP tool |
| `load_template` | `template` (required): `hello-gpu` | Overwrites workspace firmware/HDL |
| `list_files` | — | Workspace paths |
| `read_file` | `path` | e.g. `firmware/main.c`, `hdl/gpu.v` |
| `write_file` | `path`, `content` | Whole file. Jail: no `..` |
| `simulate` | — | Reset + run up to 5M cycles or `ebreak` halt. **One job.** |
| `step` | — | One clock |
| `reset` | — | Clear sim state |
| `get_uart` | — | UART log text |
| `get_leds` | — | `0–255` |
| `get_servos` | — | PWM duties `[a,b,c,d]` 0–255 |
| `get_framebuffer` | — | 4096-byte RGB332, base64 |
| `get_waves` | — | Last ≤1024 samples |
| `gate_count` | — | Yosys; errors if yosys missing |
| `export_fpga` | — | iCE40 hx8k `.bin`; slow; skip unless asked |

## Firmware contract

- `#include "tgb.h"` — use `LEDS`, `BTN`, `uart_puts`, `PWM0–3`, `gpu_*`.
- RV32I only. No libc. `main` returns → CPU **halts** (needed to finish Run).
- **Bounded loops.** A `while(1)` or huge `delay()` hits the cycle cap and looks hung.
- Skip `delay(10000)` in custom code unless you want a long sim.
- PWM duty **0–255**. LEDs **bits 0–7**.
- GPU coords **0–63**. Color **RGB332** (`RRRGGGBB`). `gpu_wait()` before the next command (helpers already wait).

```c
#include "tgb.h"

int main(void) {
    PWM0 = 10u; PWM1 = 20u; PWM2 = 30u; PWM3 = 40u;
    uart_puts("custom lab2 program\n");
    LEDS = 0x55u;
    gpu_clear(0x02u);
    gpu_wait();
    gpu_set_color(0xE0u);
    gpu_fill_rect(8, 8, 48, 48);
    gpu_wait();
    return 0;
}
```

### Address map (do not invent another)

| Device | Address |
|--------|---------|
| RAM | `0x0000_0000` |
| LEDs | `0x1000_0000` |
| BTN | `0x1000_0004` |
| UART TX | `0x2000_0000` |
| UART status | `0x2000_0004` bit0 TX ready |
| PWM n | `0x3000_0000 + 4*n` |
| GPU CMD | `0x4000_0000` |
| GPU STAT | `0x4000_0004` bit0 busy |

GPU CMD word: `[31:28]` op `1` CLEAR (color `[7:0]`), `2` SET_COLOR, `3` PUT_PIXEL (`[13:8]`=x, `[5:0]`=y), `4` FILL_RECT (`[27:22]`=x, `[21:16]`=y, `[13:8]`=w, `[5:0]`=h). Prefer `tgb.h` helpers.

## Best practices

- **Status first.** Never parallel `simulate` + `simulate`.
- **Write the whole `main.c`.** Partial edits via MCP are error-prone.
- **C before Verilog.** Students’ “blink / print / draw” tasks are firmware.
- **Prove it with dumps.** Custom success is **RAN**, not the hello-gpu PASS pill.
- **Short programs.** No chaser delays unless demonstrating LEDs over time.
- **Restore the template** on a shared Lab2 when the user is done.
- **Don’t mock PASS.** If simulate failed to build, say so and paste the log.
- **Don’t tunnel MCP** to the public internet (no auth).
- **Don’t** treat this GPU as NVIDIA/CUDA or add Cadence/TSMC talk.
- **gate_count / export_fpga** only on request; they need Yosys/nextpnr and take minutes.
- If `structuredContent` is null, parse the tool’s **text** JSON anyway.

## Exercises (run these when asked to “try the lab”)

1. **Hello** — `load_template` + `simulate`. Confirm UART hello, LEDs 170, PWM 32/96/160/224, PASS.
2. **Custom hello** — `write_file` a tiny `main.c` that `uart_puts` a unique string and sets `LEDS = 0x55`. `simulate`. Confirm UART/LEDs. Expect **RAN**, not FAIL.
3. **Rectangle** — `gpu_clear` + `gpu_fill_rect`. `get_framebuffer` non-empty / not the hello-gpu hash.
4. **Hang** — explain (don’t ship) that `while(1);` will FAIL/timeout. Always `return 0` from `main`.
5. **Restore** — `load_template` `hello-gpu` again.

## What this is not

Not Cadence, not TSMC PDKs, not CUDA, not analog, not USB to a real board (Export FPGA is a `.bin` download only). Servos in the UI are **drawn**.
