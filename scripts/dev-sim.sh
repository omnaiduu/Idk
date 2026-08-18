#!/usr/bin/env bash
set -euo pipefail

# Usage:
#   dev-sim.sh build <workspace>   — build firmware + sim into workspace/build/
#   dev-sim.sh run <workspace>     — smoke test in workspace
#   dev-sim.sh native              — full PASS check from repo root (CI / Docker)

REPO_ROOT="${REPO_ROOT:-$(cd "$(dirname "$0")/.." && pwd)}"
MODE="${1:-native}"
WORKSPACE="${2:-${REPO_ROOT}/workspace-data}"

MAIN_SRC="${WORKSPACE}/firmware/main.c"
if [[ ! -f "$MAIN_SRC" ]]; then
  MAIN_SRC="${REPO_ROOT}/templates/hello-gpu/main.c"
fi

GCC=riscv64-unknown-elf-gcc
OBJCOPY=riscv64-unknown-elf-objcopy
GCCFLAGS=(-march=rv32i -mabi=ilp32 -nostdlib -nostartfiles -ffreestanding -fno-builtin -Os
  -T "${REPO_ROOT}/firmware/link.ld" -I "${REPO_ROOT}/firmware")

build_firmware() {
  local out_dir="$1"
  mkdir -p "${out_dir}"
  echo "[dev-sim] firmware <= ${MAIN_SRC}"
  "${GCC}" "${GCCFLAGS[@]}" \
    "${REPO_ROOT}/firmware/start.S" "${MAIN_SRC}" \
    -o "${out_dir}/firmware.elf"
  "${OBJCOPY}" -O verilog "${out_dir}/firmware.elf" "${out_dir}/firmware.hex"
}

build_verilator() {
  local lock="${REPO_ROOT}/build/.verilator.lock"
  local hdl_dir="${HDL_DIR:-${WORKSPACE}/hdl}"
  if [[ ! -f "${hdl_dir}/tiny_gpu_top.v" ]]; then
    hdl_dir="${REPO_ROOT}/hdl"
  fi
  mkdir -p "${REPO_ROOT}/build"
  echo "[dev-sim] verilator (HDL_DIR=${hdl_dir})"
  (
    flock -x 200
    # Let make decide staleness — do not skip just because the binary exists.
    make -C "${REPO_ROOT}/sim" sim ROOT="${REPO_ROOT}" HDL_DIR="${hdl_dir}" MAIN_SRC="${MAIN_SRC}"
  ) 200>"$lock"
}

copy_sim_to_workspace() {
  local ws_build="$1"
  mkdir -p "${ws_build}"
  local dest="${ws_build}/sim"
  if [[ -f "$dest" ]] && cmp -s "${REPO_ROOT}/sim/obj_dir/Vtiny_gpu_top" "$dest" 2>/dev/null; then
    echo "[dev-sim] sim binary unchanged"
    return 0
  fi
  cp "${REPO_ROOT}/sim/obj_dir/Vtiny_gpu_top" "${dest}.new"
  chmod +x "${dest}.new"
  mv -f "${dest}.new" "${dest}"
}

check_pass() {
  local expected="$1"
  local out_dir="$2"
  python3 - "$expected" "$out_dir" <<'PY'
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
if exp.get("uart_contains") and exp["uart_contains"] not in uart:
    errors.append(f"UART missing {exp['uart_contains']!r}")
if exp.get("leds_nonzero") and not leds.get("leds"):
    errors.append("LEDs are zero at end")
if exp.get("leds_final") and leds.get("leds") != exp["leds_final"]:
    errors.append(f"LEDs expected {exp['leds_final']} got {leds.get('leds')}")
if exp.get("pwm_duty") and pwm.get("duty") != exp["pwm_duty"]:
    errors.append(f"PWM duty mismatch: {pwm.get('duty')}")
digest = hashlib.sha256(fb).hexdigest()
if exp.get("fb_sha256") and digest != exp["fb_sha256"]:
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

run_native() {
  mkdir -p "${REPO_ROOT}/build/sim-out" "${REPO_ROOT}/build"
  MAIN_SRC="${REPO_ROOT}/templates/hello-gpu/main.c"
  build_firmware "${REPO_ROOT}/build"
  build_verilator
  export DUMP_DIR="${REPO_ROOT}/build/sim-out"
  printf '{"cmd":"reset"}\n{"cmd":"run","cycles":5000000}\n' \
    | "${REPO_ROOT}/sim/obj_dir/Vtiny_gpu_top" "+firmware=${REPO_ROOT}/build/firmware.hex" \
    > "${REPO_ROOT}/build/sim-out/last-snapshot.json"
  check_pass "${REPO_ROOT}/templates/hello-gpu/expected.json" "${REPO_ROOT}/build/sim-out"
}

case "$MODE" in
  build)
    mkdir -p "${WORKSPACE}/dumps" "${WORKSPACE}/firmware"
    build_firmware "${WORKSPACE}/build"
    build_verilator
    copy_sim_to_workspace "${WORKSPACE}/build"
    ;;
  run)
    mkdir -p "${WORKSPACE}/dumps"
    build_firmware "${WORKSPACE}/build"
    build_verilator
    copy_sim_to_workspace "${WORKSPACE}/build"
    export DUMP_DIR="${WORKSPACE}/dumps"
    printf '{"cmd":"reset"}\n{"cmd":"run","cycles":5000000}\n' \
      | "${WORKSPACE}/build/sim" "+firmware=${WORKSPACE}/build/firmware.hex"
    ;;
  native)
    if command -v docker >/dev/null 2>&1 && docker info >/dev/null 2>&1 \
       && [[ "${FORCE_NATIVE:-}" != "1" ]] \
       && ! command -v verilator >/dev/null 2>&1; then
      echo "==> Running inside Docker (ubuntu:24.04)..."
      docker run --rm \
        -v "$REPO_ROOT:/workspace" \
        -w /workspace \
        ubuntu:24.04 \
        bash -lc '
          set -euo pipefail
          export DEBIAN_FRONTEND=noninteractive
          apt-get update -qq
          apt-get install -y -qq verilator gcc-riscv64-unknown-elf binutils-riscv64-unknown-elf make g++ python3 ca-certificates
          FORCE_NATIVE=1 ./scripts/dev-sim.sh native
        '
    else
      run_native
    fi
    ;;
  *)
    echo "unknown mode: $MODE" >&2
    exit 1
    ;;
esac
