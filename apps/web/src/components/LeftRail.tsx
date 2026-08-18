import { motion } from "framer-motion";
import { cn } from "@/lib/utils";
import { BLOCKS, getMotionDuration, type BlockId } from "@/lib/utils";
import { useReducedMotion } from "@/hooks/use-reduced-motion";

interface LeftRailProps {
  selected: BlockId;
  onSelect: (id: BlockId) => void;
}

export function LeftRail({ selected, onSelect }: LeftRailProps) {
  const reduced = useReducedMotion();

  return (
    <aside className="flex w-[200px] shrink-0 flex-col gap-1 border-r border-cream-border bg-void-panel p-2">
      <p className="px-2 py-1 text-[10px] font-medium uppercase tracking-widest text-cream-muted">
        Blocks
      </p>
      {BLOCKS.map((block) => {
        const active = selected === block.id;
        return (
          <motion.button
            key={block.id}
            type="button"
            onClick={() => onSelect(block.id)}
            aria-pressed={active}
            className={cn(
              "relative flex w-full flex-col items-start px-2 py-2 text-left transition-colors",
              active
                ? "bg-ember/5 text-ember hairline-ember"
                : "text-cream hover:bg-cream/5",
            )}
            whileHover={reduced ? undefined : { x: 2 }}
            transition={{ duration: getMotionDuration(reduced, 0.15) }}
          >
            <span className={cn("text-sm font-medium", active ? "text-ember" : "text-cream")}>
              {block.label}
            </span>
            <span className="mt-0.5 truncate font-mono text-[10px] text-cream-muted">
              {block.file}
            </span>
          </motion.button>
        );
      })}
    </aside>
  );
}
