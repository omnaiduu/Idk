import { motion, AnimatePresence } from "framer-motion";
import { Button } from "@/components/ui/button";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
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

  return (
    <Dialog open={open} onOpenChange={(v) => !v && finish()}>
      <DialogContent className="border-cream-border-strong bg-void-elevated sm:max-w-md">
        <DialogHeader>
          <DialogTitle>Welcome to the bench</DialogTitle>
          <DialogDescription>
            Step {step + 1} of {STEPS.length} — you can Run anytime.
          </DialogDescription>
        </DialogHeader>
        <AnimatePresence mode="wait">
          <motion.div
            key={step}
            initial={reduced ? false : { opacity: 0, x: 12 }}
            animate={{ opacity: 1, x: 0 }}
            exit={reduced ? undefined : { opacity: 0, x: -12 }}
            transition={{ duration: getMotionDuration(reduced, 0.2) }}
            className="space-y-3 py-2"
          >
            <div className="flex h-10 w-10 items-center justify-center rounded-md border border-ember/40 bg-ember/10">
              <Icon className="h-5 w-5 text-ember" />
            </div>
            <h3 className="text-base font-medium text-cream">{current.title}</h3>
            <p className="text-sm leading-relaxed text-cream-muted">{current.body}</p>
          </motion.div>
        </AnimatePresence>
        <DialogFooter className="gap-2 sm:gap-0">
          <Button variant="ghost" onClick={finish}>
            Skip
          </Button>
          <Button variant="ember" onClick={next}>
            {step >= STEPS.length - 1 ? "Start benching" : "Next"}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}
