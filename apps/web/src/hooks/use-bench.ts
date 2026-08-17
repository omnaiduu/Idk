import { useCallback, useEffect, useState } from "react";
import * as api from "@/api/client";
import type { ApiError, BenchStatus, McpDrawerState, PassState } from "@/types";

export function useBench() {
  const [status, setStatus] = useState<BenchStatus | null>(null);
  const [loading, setLoading] = useState(true);
  const [running, setRunning] = useState(false);
  const [passState, setPassState] = useState<PassState>("idle");
  const [cycles, setCycles] = useState(0);
  const [error, setError] = useState<ApiError | null>(null);
  const [mcpState, setMcpState] = useState<McpDrawerState | null>(null);
  const [drawerOpen, setDrawerOpen] = useState(false);

  const refreshStatus = useCallback(async () => {
    try {
      const s = await api.getStatus();
      setStatus(s);
      setCycles(s.cycles ?? 0);
      if (s.running) {
        setPassState("running");
        setRunning(true);
      } else if (s.pass === true) {
        setPassState("pass");
        setRunning(false);
      } else if (s.pass === false) {
        setPassState("fail");
        setRunning(false);
      } else {
        setPassState("idle");
        setRunning(false);
      }
      if (s.last_mcp_tool) {
        setMcpState({
          tool: s.last_mcp_tool,
          pass: s.last_mcp_pass ?? null,
          timestamp: Date.now(),
        });
      }
    } catch {
      // Backend may not be up during dev
    }
  }, []);

  useEffect(() => {
    let mounted = true;
    (async () => {
      try {
        await api.getHealth().catch(() => api.loadTemplate("hello-gpu").catch(() => undefined));
        await refreshStatus();
      } finally {
        if (mounted) setLoading(false);
      }
    })();
    return () => {
      mounted = false;
    };
  }, [refreshStatus]);

  const handleRun = useCallback(async () => {
    setError(null);
    setRunning(true);
    setPassState("running");
    try {
      const result = await api.simulate();
      setCycles(result.cycles ?? 0);
      if (result.pass === true) {
        setPassState("pass");
      } else if (result.pass === false) {
        setPassState("fail");
        setError({
          error: result.error ?? "Simulation failed",
          message: result.message,
          log: result.log,
        });
      } else {
        setPassState(result.ok ? "pass" : "fail");
      }
      await refreshStatus();
      return result;
    } catch (e) {
      const err = e as ApiError;
      setPassState("fail");
      setError(err);
      throw err;
    } finally {
      setRunning(false);
    }
  }, [refreshStatus]);

  const handleStep = useCallback(async () => {
    setError(null);
    try {
      const result = await api.step();
      setCycles(result.cycles ?? cycles);
      await refreshStatus();
      return result;
    } catch (e) {
      setError(e as ApiError);
      throw e;
    }
  }, [cycles, refreshStatus]);

  const handleReset = useCallback(async () => {
    setError(null);
    try {
      await api.reset();
      setPassState("idle");
      setCycles(0);
      await refreshStatus();
    } catch (e) {
      setError(e as ApiError);
    }
  }, [refreshStatus]);

  return {
    status,
    loading,
    running,
    passState,
    cycles,
    error,
    setError,
    mcpState,
    drawerOpen,
    setDrawerOpen,
    refreshStatus,
    handleRun,
    handleStep,
    handleReset,
  };
}
