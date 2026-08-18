import { Badge } from "@/components/ui/badge";
import {
  Sheet,
  SheetContent,
  SheetDescription,
  SheetHeader,
  SheetTitle,
} from "@/components/ui/sheet";
import { Separator } from "@/components/ui/separator";
import type { McpDrawerState } from "@/types";

interface McpDrawerProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  state: McpDrawerState | null;
}

export function McpDrawer({ open, onOpenChange, state }: McpDrawerProps) {
  const passVariant =
    state?.pass === true
      ? "pass"
      : state?.pass === false
        ? "fail"
        : state?.outcome === "ran"
          ? "ran"
          : "idle";
  const passLabel =
    state?.pass === true
      ? "PASS"
      : state?.pass === false
        ? "FAIL"
        : state?.outcome === "ran"
          ? "RAN"
          : "—";

  return (
    <Sheet open={open} onOpenChange={onOpenChange}>
      <SheetContent side="right" className="border-cream-border bg-void-elevated">
        <SheetHeader>
          <SheetTitle>MCP activity</SheetTitle>
          <SheetDescription>
            Last tool call from Cursor — not a chat panel.
          </SheetDescription>
        </SheetHeader>
        <Separator className="my-4" />
        {state ? (
          <div className="space-y-4">
            <div>
              <p className="text-[10px] uppercase tracking-widest text-cream-muted">Tool</p>
              <p className="mt-1 font-mono text-sm text-cream">{state.tool}</p>
            </div>
            <div>
              <p className="text-[10px] uppercase tracking-widest text-cream-muted">Result</p>
              <Badge variant={passVariant} className="mt-2">
                {passLabel}
              </Badge>
            </div>
            <p className="text-xs text-cream-muted">
              Connect at{" "}
              <code className="rounded bg-cream/5 px-1 py-0.5 font-mono text-[11px]">
                http://127.0.0.1:8741/mcp
              </code>
            </p>
          </div>
        ) : (
          <p className="text-sm text-cream-muted">
            No MCP tool calls yet. Paste the MCP URL in Cursor settings.
          </p>
        )}
      </SheetContent>
    </Sheet>
  );
}
