import type {
  BenchStatus,
  FileContent,
  FilesListResult,
  GateCountResult,
  LedsState,
  ServosState,
  SimulateResult,
  WavesState,
} from "@/types";

const JSON_HEADERS = { "Content-Type": "application/json" };

async function handleResponse<T>(res: Response): Promise<T> {
  const text = await res.text();
  let data: unknown = {};
  if (text) {
    try {
      data = JSON.parse(text);
    } catch {
      data = { error: text };
    }
  }

  if (!res.ok) {
    const err = data as { error?: string; message?: string; log?: string };
    throw {
      error: err.error ?? res.statusText,
      message: err.message,
      log: err.log ?? text,
      status: res.status,
    };
  }

  return data as T;
}

export async function getStatus(): Promise<BenchStatus> {
  const res = await fetch("/api/status");
  return handleResponse<BenchStatus>(res);
}

export async function getHealth(): Promise<{ ok: boolean }> {
  const res = await fetch("/health");
  return handleResponse<{ ok: boolean }>(res);
}

export async function simulate(): Promise<SimulateResult> {
  const res = await fetch("/api/simulate", { method: "POST" });
  return handleResponse<SimulateResult>(res);
}

export async function step(): Promise<SimulateResult> {
  const res = await fetch("/api/step", { method: "POST" });
  return handleResponse<SimulateResult>(res);
}

export async function reset(): Promise<SimulateResult> {
  const res = await fetch("/api/reset", { method: "POST" });
  return handleResponse<SimulateResult>(res);
}

export async function setButton(down: boolean): Promise<{ ok: boolean }> {
  const res = await fetch("/api/button", {
    method: "POST",
    headers: JSON_HEADERS,
    body: JSON.stringify({ down }),
  });
  return handleResponse<{ ok: boolean }>(res);
}

export async function gateCount(): Promise<GateCountResult> {
  const res = await fetch("/api/gate-count", { method: "POST" });
  return handleResponse<GateCountResult>(res);
}

export async function exportFpga(): Promise<Blob> {
  const res = await fetch("/api/export-fpga", { method: "POST" });
  if (!res.ok) {
    const text = await res.text();
    let data: { error?: string; message?: string; log?: string } = {};
    try {
      data = JSON.parse(text);
    } catch {
      data = { error: text };
    }
    throw {
      error: data.error ?? res.statusText,
      message: data.message,
      log: data.log ?? text,
      status: res.status,
    };
  }
  return res.blob();
}

export async function loadTemplate(name = "hello-gpu"): Promise<{ ok: boolean }> {
  const res = await fetch("/api/load-template", {
    method: "POST",
    headers: JSON_HEADERS,
    body: JSON.stringify({ template: name }),
  });
  return handleResponse<{ ok: boolean }>(res);
}

export async function listFiles(): Promise<FilesListResult> {
  const res = await fetch("/api/files");
  return handleResponse<FilesListResult>(res);
}

export async function readFile(path: string): Promise<FileContent> {
  const res = await fetch(`/api/files/${encodeURIComponent(path)}`);
  return handleResponse<FileContent>(res);
}

export async function writeFile(path: string, content: string): Promise<{ ok: boolean }> {
  const res = await fetch(`/api/files/${encodeURIComponent(path)}`, {
    method: "PUT",
    headers: JSON_HEADERS,
    body: JSON.stringify({ content }),
  });
  return handleResponse<{ ok: boolean }>(res);
}

export async function getUart(): Promise<{ text: string }> {
  const res = await fetch("/api/uart");
  return handleResponse<{ text: string }>(res);
}

export async function getLeds(): Promise<LedsState> {
  const res = await fetch("/api/leds");
  return handleResponse<LedsState>(res);
}

export async function getServos(): Promise<ServosState> {
  const res = await fetch("/api/servos");
  return handleResponse<ServosState>(res);
}

export async function getWaves(): Promise<WavesState> {
  const res = await fetch("/api/waves");
  return handleResponse<WavesState>(res);
}

export async function getFramebufferUrl(): Promise<string> {
  return `/api/framebuffer.png?t=${Date.now()}`;
}

export async function getFramebufferBinary(): Promise<ArrayBuffer> {
  const res = await fetch("/api/framebuffer");
  if (!res.ok) throw new Error("Failed to load framebuffer");
  return res.arrayBuffer();
}

export async function pollOutputs(): Promise<{
  uart: string;
  leds: LedsState;
  servos: ServosState;
  waves: WavesState;
}> {
  const [uartRes, leds, servos, waves] = await Promise.all([
    getUart().catch(() => ({ text: "" })),
    getLeds().catch(() => ({ leds: 0 })),
    getServos().catch(() => ({ duty: [0, 0, 0, 0] as [number, number, number, number] })),
    getWaves().catch(() => ({ clk: [], gpu_busy: [], led0: [], pwm0: [] })),
  ]);
  return { uart: uartRes.text, leds, servos, waves };
}
