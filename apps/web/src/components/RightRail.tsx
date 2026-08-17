import { PixelDisplay } from "@/components/PixelDisplay";
import { LedPanel } from "@/components/LedPanel";
import { ServoArms } from "@/components/ServoArms";
import { Separator } from "@/components/ui/separator";
import { Toggle } from "@/components/ui/toggle";
import type { ServosState } from "@/types";

interface RightRailProps {
  leds: number;
  servos: ServosState;
  refreshKey: number;
  hasRun: boolean;
  buttonDown: boolean;
  onButtonToggle: (down: boolean) => void;
}

export function RightRail({
  leds,
  servos,
  refreshKey,
  hasRun,
  buttonDown,
  onButtonToggle,
}: RightRailProps) {
  return (
    <aside className="flex w-[280px] shrink-0 flex-col gap-4 overflow-y-auto border-l border-cream-border bg-void-panel p-3">
      <PixelDisplay refreshKey={refreshKey} empty={!hasRun} />
      <Separator />
      <LedPanel leds={leds} />
      <Separator />
      <div className="flex items-center justify-between">
        <span className="text-[10px] font-medium uppercase tracking-widest text-cream-muted">
          Button
        </span>
        <Toggle
          pressed={buttonDown}
          onPressedChange={onButtonToggle}
          size="sm"
          variant="outline"
          aria-label="Simulated button"
        >
          {buttonDown ? "Down" : "Up"}
        </Toggle>
      </div>
      <Separator />
      <ServoArms duty={servos.duty} />
    </aside>
  );
}
