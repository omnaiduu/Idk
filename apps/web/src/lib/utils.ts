import { clsx, type ClassValue } from "clsx";
import { twMerge } from "tailwind-merge";

export function cn(...inputs: ClassValue[]) {
  return twMerge(clsx(inputs));
}

export function dutyToDegrees(duty: number): number {
  return Math.round((Math.max(0, Math.min(255, duty)) / 255) * 180);
}

export function rgb332ToCss(byte: number): string {
  const r = ((byte >> 5) & 0x07) * 36;
  const g = ((byte >> 2) & 0x07) * 36;
  const b = (byte & 0x03) * 85;
  return `rgb(${r}, ${g}, ${b})`;
}

export function formatCycles(n: number): string {
  return n.toLocaleString("en-US");
}

export const ONBOARDING_KEY = "tiny-gpu-bench-onboarding-v1";

export const BLOCKS = [
  { id: "cpu", label: "CPU", file: "firmware/main.c" },
  { id: "ram", label: "RAM", file: "hdl/ram.v" },
  { id: "gpu", label: "GPU", file: "hdl/gpu.v" },
  { id: "uart", label: "UART", file: "hdl/uart_tx.v" },
  { id: "gpio", label: "GPIO", file: "hdl/gpio_led.v" },
  { id: "pwm", label: "PWM", file: "hdl/pwm.v" },
] as const;

export type BlockId = (typeof BLOCKS)[number]["id"];

export const MOTION_FAST = 0.15;
export const MOTION_NORMAL = 0.2;
export const MOTION_SLOW = 0.25;

export function getMotionDuration(reduced: boolean, seconds = MOTION_NORMAL): number {
  return reduced ? 0 : seconds;
}
