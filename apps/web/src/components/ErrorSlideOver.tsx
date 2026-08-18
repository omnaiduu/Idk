import { motion, AnimatePresence } from "framer-motion";
import { Button } from "@/components/ui/button";
import { ScrollArea } from "@/components/ui/scroll-area";
import { getMotionDuration } from "@/lib/utils";
import { useReducedMotion } from "@/hooks/use-reduced-motion";
import type { ApiError } from "@/types";
import { Copy, X } from "lucide-react";
import { useCallback, useState } from "react";

interface ErrorSlideOverProps {
  error: ApiError | null;
  onDismiss: () => void;
}

export function ErrorSlideOver({ error, onDismiss }: ErrorSlideOverProps) {
  const reduced = useReducedMotion();
  const [copied, setCopied] = useState(false);

  const logText = [error?.error, error?.message, error?.log].filter(Boolean).join("\n\n");
  const busy = error?.status === 409;

  const copy = useCallback(async () => {
    await navigator.clipboard.writeText(logText);
    setCopied(true);
    setTimeout(() => setCopied(false), 2000);
  }, [logText]);

  return (
    <AnimatePresence>
      {error && (
        <>
          <motion.div
            className="fixed inset-0 z-40 bg-void/50"
            initial={{ opacity: 0 }}
            animate={{ opacity: 1 }}
            exit={{ opacity: 0 }}
            transition={{ duration: getMotionDuration(reduced, 0.2) }}
            onClick={onDismiss}
          />
          <motion.aside
            className="fixed bottom-0 right-0 top-12 z-50 flex w-full max-w-lg flex-col overflow-hidden border-l border-cream-border bg-void-elevated shadow-2xl"
            initial={reduced ? false : { x: "100%" }}
            animate={{ x: 0 }}
            exit={reduced ? undefined : { x: "100%" }}
            transition={{ duration: getMotionDuration(reduced, 0.25), ease: "easeOut" }}
            role="alertdialog"
            aria-labelledby="error-title"
          >
            <div className="flex shrink-0 items-center justify-between border-b border-cream-border px-4 py-3">
              <div>
                <h2 id="error-title" className="text-sm font-semibold text-cream">
                  {busy ? "Bench busy" : error.title ?? "Error"}
                </h2>
                <p className="text-xs text-cream-muted">
                  {busy ? "Wait for the current run to finish." : error.error}
                </p>
              </div>
              <div className="flex gap-1">
                <Button variant="outline" size="sm" onClick={copy}>
                  <Copy className="h-3.5 w-3.5" />
                  {copied ? "Copied" : "Copy"}
                </Button>
                <Button variant="ghost" size="icon" onClick={onDismiss}>
                  <X className="h-4 w-4" />
                </Button>
              </div>
            </div>
            <ScrollArea className="min-h-0 flex-1">
              <pre className="whitespace-pre-wrap break-words p-4 font-mono text-xs leading-relaxed text-cream">
                {logText}
              </pre>
            </ScrollArea>
          </motion.aside>
        </>
      )}
    </AnimatePresence>
  );
}
