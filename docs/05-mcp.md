# MCP — Model Context Protocol

Tiny GPU Bench exposes MCP over **Streamable HTTP** at:

```
http://127.0.0.1:8741/mcp
```

- **No authentication** — no Bearer token, no query password
- **No TLS** — localhost only
- **Origin check** — same allowed origins as the REST API (`http://127.0.0.1:8741`, `http://localhost:8741`)

## Cursor configuration

```json
{
  "mcpServers": {
    "tiny-gpu-bench": {
      "url": "http://127.0.0.1:8741/mcp"
    }
  }
}
```

## curl — initialize

```bash
curl -s -X POST http://127.0.0.1:8741/mcp \
  -H 'Content-Type: application/json' \
  -H 'Accept: application/json, text/event-stream' \
  -H 'Origin: http://127.0.0.1:8741' \
  -d '{
    "jsonrpc": "2.0",
    "id": 1,
    "method": "initialize",
    "params": {
      "protocolVersion": "2025-11-25",
      "capabilities": {},
      "clientInfo": { "name": "curl", "version": "1.0" }
    }
  }'
```

## curl — call `simulate`

After initialize (and session setup if required by your client), call the simulate tool:

```bash
curl -s -X POST http://127.0.0.1:8741/mcp \
  -H 'Content-Type: application/json' \
  -H 'Accept: application/json, text/event-stream' \
  -H 'Origin: http://127.0.0.1:8741' \
  -d '{
    "jsonrpc": "2.0",
    "id": 2,
    "method": "tools/call",
    "params": {
      "name": "simulate",
      "arguments": {}
    }
  }'
```

The response includes cycle count, PASS/FAIL, and UART excerpt — same logic as `POST /api/simulate`.

Tool results always include a JSON **object** in `structuredContent` (never `null`). Cursor Cloud rejects `structuredContent: null`.

## Tools

| Tool | REST equivalent |
|------|-----------------|
| `simulate` | `POST /api/simulate` |
| `step` | `POST /api/step` |
| `reset` | `POST /api/reset` |
| `list_files` | `GET /api/files` |
| `read_file` | `GET /api/files/{path}` |
| `write_file` | `PUT /api/files/{path}` |
| `gate_count` | `POST /api/gate-count` |
| `export_fpga` | `POST /api/export-fpga` |
| `get_uart` | `GET /api/uart` |
| `get_leds` | `GET /api/leds` |
| `get_servos` | `GET /api/servos` |
| `get_framebuffer` | `GET /api/framebuffer` |
| `get_waves` | `GET /api/waves` |
| `get_status` | `GET /api/status` |
| `load_template` | `POST /api/load-template` |

MCP handlers call the same Go `bench` package as HTTP — no duplicate sim logic.

## Security reminder

Never publish `8741` on `0.0.0.0` or through a public tunnel. Unauthenticated MCP on the open internet would let anyone run arbitrary Verilator builds in your container.
