import { useEffect, useRef, useState } from "react";
import { getFramebufferBinary } from "@/api/client";
import { rgb332ToCss } from "@/lib/utils";

const FB_SIZE = 64;
const SCALE = 8;

interface PixelDisplayProps {
  refreshKey: number;
  empty: boolean;
}

export function PixelDisplay({ refreshKey, empty }: PixelDisplayProps) {
  const canvasRef = useRef<HTMLCanvasElement>(null);
  const [error, setError] = useState(false);

  useEffect(() => {
    const canvas = canvasRef.current;
    if (!canvas) return;
    const ctx = canvas.getContext("2d");
    if (!ctx) return;

    if (empty) {
      ctx.fillStyle = "#0a0906";
      ctx.fillRect(0, 0, FB_SIZE * SCALE, FB_SIZE * SCALE);
      return;
    }

    let cancelled = false;

    (async () => {
      try {
        const buf = await getFramebufferBinary();
        if (cancelled) return;
        const data = new Uint8Array(buf);
        const imageData = ctx.createImageData(FB_SIZE, FB_SIZE);
        for (let y = 0; y < FB_SIZE; y++) {
          for (let x = 0; x < FB_SIZE; x++) {
            const byte = data[y * FB_SIZE + x] ?? 0;
            const css = rgb332ToCss(byte);
            const match = css.match(/\d+/g);
            const r = Number(match?.[0] ?? 0);
            const g = Number(match?.[1] ?? 0);
            const b = Number(match?.[2] ?? 0);
            const i = (y * FB_SIZE + x) * 4;
            imageData.data[i] = r;
            imageData.data[i + 1] = g;
            imageData.data[i + 2] = b;
            imageData.data[i + 3] = 255;
          }
        }
        const offscreen = document.createElement("canvas");
        offscreen.width = FB_SIZE;
        offscreen.height = FB_SIZE;
        offscreen.getContext("2d")?.putImageData(imageData, 0, 0);
        ctx.imageSmoothingEnabled = false;
        ctx.fillStyle = "#0a0906";
        ctx.fillRect(0, 0, FB_SIZE * SCALE, FB_SIZE * SCALE);
        ctx.drawImage(offscreen, 0, 0, FB_SIZE * SCALE, FB_SIZE * SCALE);
        setError(false);
      } catch {
        if (!cancelled) setError(true);
      }
    })();

    return () => {
      cancelled = true;
    };
  }, [refreshKey, empty]);

  return (
    <div className="relative mx-auto">
      <div className="rounded-md border-2 border-ember/70 bg-void p-1.5 glow-ember crt-inner">
        <canvas
          ref={canvasRef}
          width={FB_SIZE * SCALE}
          height={FB_SIZE * SCALE}
          className="pixelated block rounded-sm bg-[#0a0906]"
          aria-label="64 by 64 GPU framebuffer"
        />
      </div>
      {error && (
        <p className="mt-2 text-center text-[10px] text-cream-muted">No framebuffer yet</p>
      )}
      <p className="mt-2 text-center text-[10px] uppercase tracking-wider text-cream-muted">
        64×64 RGB332 · ×{SCALE}
      </p>
    </div>
  );
}
