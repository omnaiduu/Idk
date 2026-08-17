package main

import (
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"

	"tiny-gpu-bench/internal/api"
	"tiny-gpu-bench/internal/bench"
	mcpsrv "tiny-gpu-bench/internal/mcp"
	"tiny-gpu-bench/internal/sim"
	"tiny-gpu-bench/internal/workspace"
)

func main() {
	host := env("HOST", "0.0.0.0")
	port := env("PORT", "8741")
	webRoot := env("WEB_ROOT", "./apps/web/dist")
	wsRoot := env("WORKSPACE", "./workspace-data")
	timeoutSec, _ := strconv.Atoi(env("SIM_TIMEOUT_SEC", "120"))

	repoRoot, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}

	templatesDir := filepath.Join(repoRoot, "templates")
	hdlDir := filepath.Join(repoRoot, "hdl")
	fwDir := filepath.Join(repoRoot, "firmware")
	scriptsDir := filepath.Join(repoRoot, "scripts")

	ws, err := workspace.New(wsRoot, templatesDir, hdlDir, fwDir)
	if err != nil {
		log.Fatal(err)
	}
	if err := ws.LoadTemplate("hello-gpu"); err != nil {
		log.Printf("warning: load template: %v", err)
	}

	runner := sim.NewRunner(repoRoot, scriptsDir, ws, templatesDir, bench.DefaultTimeout(timeoutSec))
	b := bench.New(ws, runner)

	apiServer := api.NewServer(b, webRoot)
	mcpServer := mcpsrv.NewServer(b)

	mux := http.NewServeMux()
	mux.Handle("/mcp", mcpServer.Handler())
	mux.Handle("/", apiServer.Handler())

	addr := host + ":" + port
	log.Printf("tiny-gpu-bench listening on %s", addr)
	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Fatal(err)
	}
}

func env(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
