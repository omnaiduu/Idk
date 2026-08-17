import { cn } from "@/lib/utils";

interface LedPanelProps {
  leds: number;
}

export function LedPanel({ leds }: LedPanelProps) {
  return (
    <div className="space-y-2">
      <p className="text-[10px] font-medium uppercase tracking-widest text-cream-muted">LEDs</p>
      <div className="flex flex-wrap justify-center gap-2">
        {Array.from({ length: 8 }).map((_, i) => {
          const on = (leds >> i) & 1;
          return (
            <div
              key={i}
              className={cn(
                "h-4 w-4 rounded-full border transition-shadow duration-200",
                on
                  ? "border-ember bg-ember glow-ember"
                  : "border-cream-border bg-cream/5",
              )}
              aria-label={`LED ${i} ${on ? "on" : "off"}`}
            />
          );
        })}
      </div>
    </div>
  );
}
