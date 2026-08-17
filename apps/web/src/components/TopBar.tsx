import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import {
  Tooltip,
  TooltipContent,
  TooltipProvider,
  TooltipTrigger,
} from "@/components/ui/tooltip";
import { formatCycles } from "@/lib/utils";
import type { PassState } from "@/types";
import { Activity, Cpu, Download, Play, RotateCcw, SkipForward, Sparkles } from "lucide-react";

interface TopBarProps {
  cycles: number;
  passState: PassState;
  running: boolean;
  onRun: () => void;
  onStep: () => void;
  onReset: () => void;
  onGateCount: () => void;
  onExportFpga: () => void;
  onOpenMcp: () => void;
}

export function TopBar({
  cycles,
  passState,
  running,
  onRun,
  onStep,
  onReset,
  onGateCount,
  onExportFpga,
  onOpenMcp,
}: TopBarProps) {
  const badgeVariant =
    passState === "pass"
      ? "pass"
      : passState === "fail"
        ? "fail"
        : passState === "running"
          ? "running"
          : "idle";

  const badgeLabel =
    passState === "pass"
      ? "PASS"
      : passState === "fail"
        ? "FAIL"
        : passState === "running"
          ? "RUN"
          : "—";

  return (
    <TooltipProvider delayDuration={300}>
      <header className="flex h-12 shrink-0 items-center gap-3 border-b border-cream-border bg-void px-4">
        <div className="flex items-center gap-2">
          <Cpu className="h-4 w-4 text-ember" strokeWidth={1.75} />
          <h1 className="text-sm font-semibold tracking-tight text-cream">Tiny GPU Bench</h1>
        </div>

        <SeparatorDot />

        <div className="flex items-center gap-1.5 text-xs text-cream-muted">
          <span>Cycles</span>
          <span className="tabular-nums font-mono text-cream">{formatCycles(cycles)}</span>
        </div>

        <Badge variant={badgeVariant} className="min-w-[52px] justify-center">
          {badgeLabel}
        </Badge>

        <div className="ml-auto flex items-center gap-1.5">
          <Tooltip>
            <TooltipTrigger asChild>
              <Button variant="ember" size="sm" onClick={onRun} disabled={running}>
                <Play className="h-3.5 w-3.5 fill-current" />
                {running ? "Running…" : "Run"}
              </Button>
            </TooltipTrigger>
            <TooltipContent>
              {running ? "Running the fake chip…" : "Reset + simulate up to 2M cycles"}
            </TooltipContent>
          </Tooltip>

          <Tooltip>
            <TooltipTrigger asChild>
              <Button variant="outline" size="sm" onClick={onStep} disabled={running}>
                <SkipForward className="h-3.5 w-3.5" />
                Step
              </Button>
            </TooltipTrigger>
            <TooltipContent>Advance one clock cycle</TooltipContent>
          </Tooltip>

          <Tooltip>
            <TooltipTrigger asChild>
              <Button variant="ghost" size="sm" onClick={onReset} disabled={running}>
                <RotateCcw className="h-3.5 w-3.5" />
                Reset
              </Button>
            </TooltipTrigger>
            <TooltipContent>Reset simulation state</TooltipContent>
          </Tooltip>

          <SeparatorDot />

          <Tooltip>
            <TooltipTrigger asChild>
              <Button variant="outline" size="sm" onClick={onGateCount} disabled={running}>
                <Sparkles className="h-3.5 w-3.5" />
                How big?
              </Button>
            </TooltipTrigger>
            <TooltipContent>Yosys gate count (stat)</TooltipContent>
          </Tooltip>

          <Tooltip>
            <TooltipTrigger asChild>
              <Button variant="outline" size="sm" onClick={onExportFpga} disabled={running}>
                <Download className="h-3.5 w-3.5" />
                Export FPGA
              </Button>
            </TooltipTrigger>
            <TooltipContent>Build iCE40 hx8k .bin</TooltipContent>
          </Tooltip>

          <Tooltip>
            <TooltipTrigger asChild>
              <Button variant="ghost" size="icon" onClick={onOpenMcp}>
                <Activity className="h-4 w-4" />
              </Button>
            </TooltipTrigger>
            <TooltipContent>Last MCP tool call</TooltipContent>
          </Tooltip>
        </div>
      </header>
    </TooltipProvider>
  );
}

function SeparatorDot() {
  return <div className="h-4 w-px bg-cream-border" aria-hidden />;
}
