import { motion, AnimatePresence } from "framer-motion";
import { Button } from "@/components/ui/button";
import { getMotionDuration } from "@/lib/utils";
import { useReducedMotion } from "@/hooks/use-reduced-motion";
import { Cpu, Monitor, Play } from "lucide-react";
import { useState } from "react";

const STEPS = [
  {
    icon: Cpu,
    title: "Fake chip in the browser",
    body: "Verilog describes the machine. Press Run and the server simulates it with Verilator — no USB, no factory tools.",
  },
  {
    icon: Play,
    title: "C is the brain",
    body: "Your firmware runs on PicoRV32 inside the design. It talks to LEDs, UART, PWM, and the GPU through memory addresses.",
  },
  {
    icon: Monitor,
    title: "GPU is the screen",
    body: "The tiny GPU paints a 64×64 framebuffer when C writes commands. Not CUDA — just pixels on our lab bench.",
  },
];

interface OnboardingOverlayProps {
  open: boolean;
  onComplete: () => void;
}

export function OnboardingOverlay({ open, onComplete }: OnboardingOverlayProps) {
  const reduced = useReducedMotion();
  const [step, setStep] = useState(0);
  const current = STEPS[step];
  const Icon = current?.icon ?? Cpu;

  const finish = () => {
    onComplete();
    setStep(0);
  };

  const next = () => {
    if (step >= STEPS.length - 1) finish();
    else setStep((s) => s + 1);
  };

  if (!open) return null;

  return (
    <div className="pointer-events-none fixed inset-x-0 bottom-0 z-30 flex justify-start p-6">
      <aside
        className="pointer-events-auto w-[360px] rounded-lg border border-cream-border-strong bg-void-elevated p-4 shadow-[0_12px_40px_rgba(0,0,0,0.45)]"
        role="dialog"
        aria-modal="false"
        aria-labelledby="onboard-title"
      >
        <p className="text-[10px] uppercase tracking-widest text-cream-muted">
          Step {step + 1} of {STEPS.length} · Run is free whenever you want
        </p>
        <AnimatePresence mode="wait">
          <motion.div
            key={step}
            initial={reduced ? false : { opacity: 0, y: 8 }}
            animate={{ opacity: 1, y: 0 }}
            exit={reduced ? undefined : { opacity: 0, y: -8 }}
            transition={{ duration: getMotionDuration(reduced, 0.2) }}
            className="mt-3 space-y-2"
          >
            <div className="flex h-9 w-9 items-center justify-center rounded-sm border border-ember/40 bg-ember/10">
              <Icon className="h-4 w-4 text-ember" />
            </div>
            <h2 id="onboard-title" className="text-sm font-medium text-cream">
              {current.title}
            </h2>
            <p className="text-sm leading-relaxed text-cream-muted">{current.body}</p>
          </motion.div>
        </AnimatePresence>
        <div className="mt-4 flex justify-end gap-2">
          <Button variant="ghost" size="sm" onClick={finish}>
            Skip
          </Button>
          <Button variant="ember" size="sm" onClick={next}>
            {step >= STEPS.length - 1 ? "Got it" : "Next"}
          </Button>
        </div>
      </aside>
    </div>
  );
}
