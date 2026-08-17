import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
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
      <Tabs defaultValue="waves" className="flex min-h-0 flex-col">
        <TabsList className="w-fit">
          <TabsTrigger value="waves">Waves</TabsTrigger>
        </TabsList>
        <TabsContent value="waves" className="mt-2 min-h-0 flex-1 data-[state=inactive]:hidden">
          <WaveCanvas waves={waves} />
        </TabsContent>
      </Tabs>
      <UartLog text={uart} empty={!hasRun} />
    </footer>
  );
}
