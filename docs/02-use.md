# How to use the bench

## First visit

1. Open [http://127.0.0.1:8741](http://127.0.0.1:8741).
2. Dismiss the **three-step overlay** (fake chip / C is the brain / GPU is the screen). It stays dismissed via `localStorage`.
3. Press **Run** (ember button, top bar).

After the sim finishes (~tens of seconds on first run):

- UART log shows `hello from tiny-gpu`
- LEDs glow (not all off)
- 64×64 screen shows rectangles and **OM**
- Servo arms rotate to match PWM duties
- PASS pill turns green

## Edit code

- **Left rail** — click CPU, RAM, GPU, UART, GPIO, or PWM to switch Monaco files (typically `firmware/main.c` or HDL).
- Edit in the center editor; changes save to the workspace on Run (or via API).

## Controls

| Button | Action |
|--------|--------|
| **Run** | Reset, then simulate up to 2M cycles or until firmware halts |
| **Step** | Advance one clock cycle (live sim process) |
| **Reset** | Reset simulation state |
| **How big?** | Yosys gate count dialog |
| **Export FPGA** | Download `.bin` for iCE40 hx8k or show error |
| **MCP drawer** | Last MCP tool call and PASS/FAIL (not a chat) |

Toggle the **button** input in the right rail; the BTN register at `0x1000_0004` reflects it.

## PASS and FAIL

PASS is numeric: `templates/hello-gpu/expected.json` checks UART substring, final LEDs, PWM duties, framebuffer SHA256, and CPU halt. If you break `main.c` (e.g. delete a semicolon), Run shows **FAIL**, an error slide-over with the full log, and **Copy** — no browser `alert`.

## Cursor MCP

Add to Cursor settings:

```json
{
  "mcpServers": {
    "tiny-gpu-bench": {
      "url": "http://127.0.0.1:8741/mcp"
    }
  }
}
```

No headers, no Bearer token. See [05-mcp.md](05-mcp.md).

## Reduced motion

The UI respects `prefers-reduced-motion: reduce` — animations become instant; the bench stays fully usable.
