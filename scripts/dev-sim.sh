#!/usr/bin/env bash
# Build and run Tiny GPU Bench Verilator sim; exit 0 on PASS.
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
cd "$ROOT"

SIM_BIN="$ROOT/sim/obj_dir/Vtiny_gpu_top"
FW_HEX="$ROOT/build/firmware.hex"
OUT_DIR="$ROOT/build/sim-out"
EXPECTED="$ROOT/templates/hello-gpu/expected.json"

check_pass() {
    python3 - "$EXPECTED" "$OUT_DIR" <<'PY'
import json, hashlib, sys

expected_path, out_dir = sys.argv[1], sys.argv[2]
with open(expected_path) as f:
    exp = json.load(f)

uart = open(f"{out_dir}/uart.log", "rb").read().decode("utf-8", errors="replace")
leds = json.load(open(f"{out_dir}/leds.json"))
pwm = json.load(open(f"{out_dir}/pwm.json"))
fb = open(f"{out_dir}/fb.bin", "rb").read()
lines = open(f"{out_dir}/last-snapshot.json").read().strip().splitlines()
snap = json.loads(lines[-1])

errors = []
if exp["uart_contains"] not in uart:
    errors.append(f"UART missing {exp['uart_contains']!r}")
if not leds.get("leds", 0):
    errors.append("LEDs are zero at end")
if exp.get("leds_final") and leds.get("leds") != exp["leds_final"]:
    errors.append(f"LEDs expected {exp['leds_final']} got {leds.get('leds')}")
if pwm.get("duty") != exp.get("pwm_duty"):
    errors.append(f"PWM duty mismatch: {pwm.get('duty')}")
digest = hashlib.sha256(fb).hexdigest()
if digest != exp["fb_sha256"]:
    errors.append(f"fb_sha256 mismatch: got {digest}")
if exp.get("halted") and not snap.get("halted"):
    errors.append("CPU did not halt")

if errors:
    print("FAIL:")
    for e in errors:
        print(" -", e)
    sys.exit(1)

print("PASS: hello-gpu simulation OK")
PY
}

run_sim() {
    echo "==> Building firmware and simulation..."
    make -C "$ROOT/sim" sim

    echo "==> Running simulation..."
    mkdir -p "$OUT_DIR"
    printf '{"cmd":"reset"}\n{"cmd":"run","cycles":5000000}\n' \
        | "$SIM_BIN" +firmware="$FW_HEX" > "$OUT_DIR/last-snapshot.json"

    check_pass
}

if [[ "${1:-}" == "native" ]]; then
    run_sim
    exit 0
fi

if command -v docker >/dev/null 2>&1 && docker info >/dev/null 2>&1; then
    echo "==> Running inside Docker (ubuntu:24.04)..."
    docker run --rm \
        -v "$ROOT:/workspace" \
        -w /workspace \
        ubuntu:24.04 \
        bash -lc '
            set -euo pipefail
            export DEBIAN_FRONTEND=noninteractive
            apt-get update -qq
            apt-get install -y -qq verilator gcc-riscv64-unknown-elf binutils-riscv64-unknown-elf make g++ python3 ca-certificates
            ./scripts/dev-sim.sh native
        '
else
    echo "==> Docker not available; running natively..."
    run_sim
fi
