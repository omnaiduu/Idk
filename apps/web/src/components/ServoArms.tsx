import { motion } from "framer-motion";
import { dutyToDegrees, getMotionDuration } from "@/lib/utils";
import { useReducedMotion } from "@/hooks/use-reduced-motion";

interface ServoArmsProps {
  duty: [number, number, number, number];
}

export function ServoArms({ duty }: ServoArmsProps) {
  const reduced = useReducedMotion();

  return (
    <div className="space-y-2">
      <p className="text-[10px] font-medium uppercase tracking-widest text-cream-muted">PWM servos</p>
      <div className="grid grid-cols-4 gap-2">
        {duty.map((d, i) => {
          const angle = dutyToDegrees(d);
          return (
            <div key={i} className="flex flex-col items-center gap-1">
              <div className="relative flex h-16 w-10 items-end justify-center rounded-md border border-cream-border bg-void-elevated">
                <motion.div
                  className="absolute bottom-2 left-1/2 h-10 w-0.5 origin-bottom bg-ember/80"
                  style={{ marginLeft: -1 }}
                  animate={{ rotate: angle - 90 }}
                  transition={{
                    duration: getMotionDuration(reduced, 0.2),
                    ease: "easeOut",
                  }}
                >
                  <div className="absolute -top-1 left-1/2 h-2 w-2 -translate-x-1/2 rounded-full bg-ember glow-ember" />
                </motion.div>
                <div className="absolute bottom-1 h-2 w-2 rounded-full bg-cream/20" />
              </div>
              <span className="font-mono text-[9px] tabular-nums text-cream-muted">{d}</span>
            </div>
          );
        })}
      </div>
      <p className="text-center text-[10px] italic text-cream-muted/80">Drawn motors — not real.</p>
    </div>
  );
}
