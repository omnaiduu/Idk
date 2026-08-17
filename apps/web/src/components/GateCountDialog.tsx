import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";

interface GateCountDialogProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  loading: boolean;
  cells?: number;
  wires?: number;
  error?: string;
}

export function GateCountDialog({
  open,
  onOpenChange,
  loading,
  cells,
  wires,
  error,
}: GateCountDialogProps) {
  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="border-cream-border bg-void-elevated sm:max-w-sm">
        <DialogHeader>
          <DialogTitle>How big?</DialogTitle>
          <DialogDescription>Yosys stat on your Verilog design.</DialogDescription>
        </DialogHeader>
        {loading ? (
          <p className="py-4 text-sm text-cream-muted">Running Yosys stat…</p>
        ) : error ? (
          <p className="py-4 text-sm text-red-300">{error}</p>
        ) : (
          <dl className="grid grid-cols-2 gap-3 py-2">
            <div className="rounded-md border border-cream-border bg-void-panel p-3">
              <dt className="text-[10px] uppercase tracking-widest text-cream-muted">Cells</dt>
              <dd className="mt-1 font-mono text-lg tabular-nums text-cream">
                {cells?.toLocaleString() ?? "—"}
              </dd>
            </div>
            <div className="rounded-md border border-cream-border bg-void-panel p-3">
              <dt className="text-[10px] uppercase tracking-widest text-cream-muted">Wires</dt>
              <dd className="mt-1 font-mono text-lg tabular-nums text-cream">
                {wires?.toLocaleString() ?? "—"}
              </dd>
            </div>
          </dl>
        )}
      </DialogContent>
    </Dialog>
  );
}
