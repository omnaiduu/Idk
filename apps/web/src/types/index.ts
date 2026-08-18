export type PassState = "idle" | "pass" | "fail" | "ran" | "running";

export interface BenchStatus {
  ok: boolean;
  cycles: number;
  pass: boolean | null;
  outcome?: "pass" | "fail" | "ran";
  running: boolean;
  busy?: boolean;
  message?: string;
  last_mcp_tool?: string;
  last_mcp_pass?: boolean | null;
}

export interface LedsState {
  leds: number;
}

export interface ServosState {
  duty: [number, number, number, number];
}

export interface WavesState {
  clk: number[];
  gpu_busy: number[];
  led0: number[];
  pwm0: number[];
}

export interface SimulateResult {
  ok: boolean;
  pass?: boolean;
  outcome?: "pass" | "fail" | "ran";
  cycles?: number;
  message?: string;
  error?: string;
  log?: string;
}

export interface GateCountResult {
  ok: boolean;
  cells?: number;
  wires?: number;
  message?: string;
  error?: string;
}

export interface FileEntry {
  path: string;
  name: string;
  is_dir?: boolean;
}

export interface FilesListResult {
  files: FileEntry[];
}

export interface FileContent {
  path: string;
  content: string;
}

export interface ApiError {
  error: string;
  title?: string;
  message?: string;
  log?: string;
  status?: number;
}

export interface McpDrawerState {
  tool: string;
  pass: boolean | null;
  outcome?: "pass" | "fail" | "ran";
  timestamp: number;
}
