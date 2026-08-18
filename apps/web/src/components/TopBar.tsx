import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Tooltip, TooltipContent, TooltipTrigger } from "@/components/ui/tooltip";
import { formatCycles } from "@/lib/utils";
import type { PassState } from "@/types";
import { Activity, Cpu, Download, Play, RotateCcw, SkipForward, Sparkles } from "lucide-react";
import type { ReactNode } from "react";

interface TopBarProps {
  cycles: number;
  passState: PassState;
  running: boolean;
  busy?: boolean;
  onRun: () => void;
  onStep: () => void;
  onReset: () => void;
  onGateCount: () => void;
  onExportFpga: () => void;
  onOpenMcp: () => void;
}

function ActionTip({
  label,
  disabled,
  children,
}: {
  label: string;
  disabled?: boolean;
  children: ReactNode;
}) {
  return (
    <Tooltip>
      <TooltipTrigger asChild>
        {disabled ? <span className="inline-flex">{children}</span> : children}
      </TooltipTrigger>
      <TooltipContent>{label}</TooltipContent>
    </Tooltip>
  );
}

export function TopBar({
  cycles,
  passState,
  running,
  busy,
  onRun,
  onStep,
  onReset,
  onGateCount,
  onExportFpga,
  onOpenMcp,
}: TopBarProps) {
  const locked = running || Boolean(busy);
  const simulating = passState === "running";
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
        <ActionTip
          disabled={locked}
          label={
            simulating
              ? "Running the fake chip…"
              : locked
                ? "Bench is busy"
                : "Reset + simulate until halt (cap 5M cycles)"
          }
        >
          <Button variant="ember" size="sm" onClick={onRun} disabled={locked} aria-label="Run simulation">
            <Play className="h-3.5 w-3.5 fill-current" />
            {simulating ? "Running the fake chip…" : "Run"}
          </Button>
        </ActionTip>

        <ActionTip disabled={locked} label="Advance one clock cycle">
          <Button variant="outline" size="sm" onClick={onStep} disabled={locked}>
            <SkipForward className="h-3.5 w-3.5" />
            Step
          </Button>
        </ActionTip>

        <ActionTip disabled={locked} label="Reset simulation state">
          <Button variant="ghost" size="sm" onClick={onReset} disabled={locked}>
            <RotateCcw className="h-3.5 w-3.5" />
            Reset
          </Button>
        </ActionTip>

        <SeparatorDot />

        <ActionTip disabled={locked} label="Yosys gate count (stat)">
          <Button variant="outline" size="sm" onClick={onGateCount} disabled={locked}>
            <Sparkles className="h-3.5 w-3.5" />
            How big?
          </Button>
        </ActionTip>

        <ActionTip disabled={locked} label="Build iCE40 hx8k .bin">
          <Button variant="outline" size="sm" onClick={onExportFpga} disabled={locked}>
            <Download className="h-3.5 w-3.5" />
            Export FPGA
          </Button>
        </ActionTip>

        <ActionTip label="Last MCP tool call">
          <Button variant="ghost" size="icon" onClick={onOpenMcp}>
            <Activity className="h-4 w-4" />
          </Button>
        </ActionTip>
      </div>
    </header>
  );
}

function SeparatorDot() {
  return <div className="h-4 w-px bg-cream-border" aria-hidden />;
}
