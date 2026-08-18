package mcp

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"tiny-gpu-bench/internal/bench"
	"tiny-gpu-bench/internal/sim"
	"tiny-gpu-bench/internal/workspace"
)

func testBench(t *testing.T) *bench.Bench {
	t.Helper()
	root := t.TempDir()
	repo, err := filepath.Abs(filepath.Join("..", ".."))
	if err != nil {
		t.Fatal(err)
	}
	ws, err := workspace.New(root, filepath.Join(repo, "templates"), filepath.Join(repo, "hdl"), filepath.Join(repo, "firmware"))
	if err != nil {
		t.Fatal(err)
	}
	runner := sim.NewRunner(repo, filepath.Join(repo, "scripts"), ws, filepath.Join(repo, "templates"), time.Second)
	return bench.New(ws, runner)
}

func TestTextResultIsJSONObject(t *testing.T) {
	res, obj, err := textResult(map[string]any{"ok": true, "leds": 170})
	if err != nil {
		t.Fatal(err)
	}
	if obj == nil {
		t.Fatal("structured output is nil")
	}
	if res.StructuredContent == nil {
		t.Fatal("CallToolResult.StructuredContent is nil")
	}
	raw, err := json.Marshal(res)
	if err != nil {
		t.Fatal(err)
	}
	var wire map[string]any
	if err := json.Unmarshal(raw, &wire); err != nil {
		t.Fatal(err)
	}
	sc, ok := wire["structuredContent"].(map[string]any)
	if !ok {
		t.Fatalf("structuredContent want object, got %T %v", wire["structuredContent"], wire["structuredContent"])
	}
	if sc["leds"] != float64(170) {
		t.Fatalf("leds=%v", sc["leds"])
	}
}

func TestErrorResultIsJSONObject(t *testing.T) {
	res, obj, err := errorResult(fmt.Errorf("nope"))
	if err != nil {
		t.Fatal(err)
	}
	if !res.IsError {
		t.Fatal("expected IsError")
	}
	if obj["error"] != "nope" {
		t.Fatalf("obj=%v", obj)
	}
	raw, err := json.Marshal(res)
	if err != nil {
		t.Fatal(err)
	}
	var wire map[string]any
	if err := json.Unmarshal(raw, &wire); err != nil {
		t.Fatal(err)
	}
	if _, ok := wire["structuredContent"].(map[string]any); !ok {
		t.Fatalf("structuredContent want object, got %T %v", wire["structuredContent"], wire["structuredContent"])
	}
}

func sseJSON(t *testing.T, body string) map[string]any {
	t.Helper()
	for _, line := range strings.Split(body, "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "data:") {
			payload := strings.TrimSpace(strings.TrimPrefix(line, "data:"))
			var msg map[string]any
			if err := json.Unmarshal([]byte(payload), &msg); err != nil {
				t.Fatalf("sse json: %v in %q", err, payload)
			}
			return msg
		}
	}
	var msg map[string]any
	if err := json.Unmarshal([]byte(body), &msg); err != nil {
		t.Fatalf("no json-rpc in body: %s", body)
	}
	return msg
}

func mcpPOST(t *testing.T, url, session, body string) (http.Header, string) {
	t.Helper()
	req, err := http.NewRequest(http.MethodPost, url, strings.NewReader(body))
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json, text/event-stream")
	if session != "" {
		req.Header.Set("Mcp-Session-Id", session)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	b, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode >= 300 {
		t.Fatalf("status %d: %s", resp.StatusCode, b)
	}
	return resp.Header, string(b)
}

func TestHTTPGetStatusStructuredContentObject(t *testing.T) {
	ts := httptest.NewServer(NewServer(testBench(t)).Handler())
	t.Cleanup(ts.Close)

	hdr, initBody := mcpPOST(t, ts.URL, "", `{
		"jsonrpc":"2.0","id":1,"method":"initialize",
		"params":{"protocolVersion":"2025-11-25","capabilities":{},"clientInfo":{"name":"test","version":"1"}}
	}`)
	initMsg := sseJSON(t, initBody)
	if initMsg["error"] != nil {
		t.Fatalf("initialize error: %v", initMsg["error"])
	}
	session := hdr.Get("Mcp-Session-Id")
	if session == "" {
		t.Fatal("missing Mcp-Session-Id")
	}
	mcpPOST(t, ts.URL, session, `{"jsonrpc":"2.0","method":"notifications/initialized"}`)

	_, callBody := mcpPOST(t, ts.URL, session, `{
		"jsonrpc":"2.0","id":2,"method":"tools/call",
		"params":{"name":"get_status","arguments":{}}
	}`)
	msg := sseJSON(t, callBody)
	result, ok := msg["result"].(map[string]any)
	if !ok {
		t.Fatalf("result: %v", msg)
	}
	sc, ok := result["structuredContent"].(map[string]any)
	if !ok {
		t.Fatalf("structuredContent want object, got %T %v", result["structuredContent"], result["structuredContent"])
	}
	if sc["ok"] != true {
		t.Fatalf("status=%v", sc)
	}
}
