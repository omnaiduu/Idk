import { WaveCanvas } from "@/components/WaveCanvas";
import { UartLog } from "@/components/UartLog";
import type { WavesState } from "@/types";

interface BottomPanelProps {
  waves: WavesState;
  uart: string;
  hasRun: boolean;
}

export function BottomPanel({ waves, uart, hasRun }: BottomPanelProps) {
  return (
    <footer className="grid h-[180px] shrink-0 grid-cols-[1fr_320px] gap-3 border-t border-cream-border bg-void-panel p-3">
      <div className="flex min-h-0 flex-col">
        <p className="mb-2 text-[10px] font-medium uppercase tracking-widest text-cream-muted">
          Waves
        </p>
        <div className="min-h-0 flex-1">
          <WaveCanvas waves={waves} />
        </div>
      </div>
      <UartLog text={uart} empty={!hasRun} />
    </footer>
  );
}
