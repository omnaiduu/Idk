import { ScrollArea } from "@/components/ui/scroll-area";

interface UartLogProps {
  text: string;
  empty: boolean;
}

export function UartLog({ text, empty }: UartLogProps) {
  const display = text.trim() || (empty ? "Press Run. You should see hello and a rectangle." : "");

  return (
    <div className="flex h-full min-h-0 flex-col rounded-md border border-cream-border bg-void">
      <div className="border-b border-cream-border px-3 py-1.5 text-[10px] font-medium uppercase tracking-widest text-cream-muted">
        UART log
      </div>
      <ScrollArea className="min-h-0 flex-1">
        <pre className="whitespace-pre-wrap break-words p-3 font-mono text-xs leading-relaxed text-cream">
          {display}
        </pre>
      </ScrollArea>
    </div>
  );
}
