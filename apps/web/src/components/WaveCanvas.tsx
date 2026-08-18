import { useEffect, useRef } from "react";
import type { WavesState } from "@/types";

const VOID = "#14120b";
const CREAM = "#efece6";
const EMBER = "#ff6a2a";
const CREAM_DIM = "rgba(239, 236, 230, 0.55)";

const CHANNELS: { key: keyof WavesState; label: string; color: string; scale?: number }[] = [
  { key: "clk", label: "clk", color: CREAM },
  { key: "gpu_busy", label: "gpu_busy", color: EMBER },
  { key: "led0", label: "led[0]", color: CREAM_DIM },
  { key: "pwm0", label: "pwm0", color: "rgba(255, 106, 42, 0.7)", scale: 255 },
];

interface WaveCanvasProps {
  waves: WavesState;
}

export function WaveCanvas({ waves }: WaveCanvasProps) {
  const canvasRef = useRef<HTMLCanvasElement>(null);

  useEffect(() => {
    const canvas = canvasRef.current;
    if (!canvas) return;

    const draw = () => {
      const ctx = canvas.getContext("2d");
      if (!ctx) return;

      const dpr = window.devicePixelRatio || 1;
      const rect = canvas.getBoundingClientRect();
      canvas.width = Math.max(1, Math.floor(rect.width * dpr));
      canvas.height = Math.max(1, Math.floor(rect.height * dpr));
      ctx.setTransform(dpr, 0, 0, dpr, 0, 0);

      const w = rect.width;
      const h = rect.height;
      const laneH = h / CHANNELS.length;

      ctx.fillStyle = VOID;
      ctx.fillRect(0, 0, w, h);

      CHANNELS.forEach((ch, lane) => {
        const data = waves[ch.key] ?? [];
        const y0 = lane * laneH;
        const mid = y0 + laneH * 0.55;
        const amp = laneH * 0.35;

        ctx.strokeStyle = "rgba(239, 236, 230, 0.08)";
        ctx.beginPath();
        ctx.moveTo(0, mid);
        ctx.lineTo(w, mid);
        ctx.stroke();

        ctx.fillStyle = "rgba(239, 236, 230, 0.45)";
        ctx.font = "10px Geist Mono, monospace";
        ctx.fillText(ch.label, 6, y0 + 12);

        if (data.length < 2) return;

        ctx.strokeStyle = ch.color;
        ctx.lineWidth = 1.25;
        ctx.beginPath();

        data.forEach((v, i) => {
          const x = (i / (data.length - 1)) * (w - 12) + 6;
          const norm = ch.scale ? v / ch.scale : v;
          const y = mid - norm * amp;
          if (i === 0) ctx.moveTo(x, y);
          else ctx.lineTo(x, y);
        });
        ctx.stroke();
      });
    };

    draw();
    const ro = new ResizeObserver(draw);
    ro.observe(canvas);
    return () => ro.disconnect();
  }, [waves]);

  return (
    <canvas
      ref={canvasRef}
      className="h-full w-full rounded-md border border-cream-border bg-void"
      aria-label="Signal waves: clk, gpu_busy, led0, pwm0"
    />
  );
}
