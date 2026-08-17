import * as React from "react";
import { cva, type VariantProps } from "class-variance-authority";
import { cn } from "@/lib/utils";

const badgeVariants = cva(
  "inline-flex items-center rounded-full border px-2 py-0.5 text-[11px] font-medium uppercase tracking-wide transition-colors",
  {
    variants: {
      variant: {
        default: "border-cream-border bg-cream/5 text-cream-muted",
        pass: "border-green-500/40 bg-green-950/40 text-green-300 glow-pass",
        fail: "border-red-500/40 bg-red-950/40 text-red-300",
        running: "border-ember/40 bg-ember/10 text-ember animate-pulse",
        idle: "border-cream-border bg-transparent text-cream-muted",
      },
    },
    defaultVariants: {
      variant: "default",
    },
  },
);

export interface BadgeProps
  extends React.HTMLAttributes<HTMLDivElement>,
    VariantProps<typeof badgeVariants> {}

function Badge({ className, variant, ...props }: BadgeProps) {
  return <div className={cn(badgeVariants({ variant }), className)} {...props} />;
}

export { Badge, badgeVariants };
