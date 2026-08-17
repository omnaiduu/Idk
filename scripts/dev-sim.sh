#!/usr/bin/env bash
set -euo pipefail

# Usage:
#   dev-sim.sh build <workspace>   — build firmware + sim into workspace/build/
#   dev-sim.sh run <workspace>     — smoke test in workspace
#   dev-sim.sh native              — full PASS check from repo root (CI)

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
  echo "[dev-sim] verilator build"
  make -C "${REPO_ROOT}/sim" sim ROOT="${REPO_ROOT}" MAIN_SRC="${MAIN_SRC}"
}

copy_sim_to_workspace() {
  local ws_build="$1"
  mkdir -p "${ws_build}"
  cp "${REPO_ROOT}/sim/obj_dir/Vtiny_gpu_top" "${ws_build}/sim"
  chmod +x "${ws_build}/sim"
}

check_pass_repo() {
  python3 - "${REPO_ROOT}/templates/hello-gpu/expected.json" "${REPO_ROOT}/build/sim-out" <<'PY'
import json, hashlib, sys
exp = json.load(open(sys.argv[1]))
out = sys.argv[2]
uart = open(f"{out}/uart.log", "rb").read().decode("utf-8", errors="replace")
leds = json.load(open(f"{out}/leds.json"))
fb = open(f"{out}/fb.bin", "rb").read()
errs = []
if exp.get("uart_contains") and exp["uart_contains"] not in uart:
    errs.append("uart")
if exp.get("leds_nonzero") and not leds.get("leds"):
    errs.append("leds")
if exp.get("fb_sha256") and hashlib.sha256(fb).hexdigest() != exp["fb_sha256"]:
    errs.append("fb_sha256")
if errs:
    print("FAIL:", ", ".join(errs)); sys.exit(1)
print("PASS")
PY
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
    printf '{"cmd":"reset"}\n{"cmd":"run","cycles":2000000}\n' \
      | "${WORKSPACE}/build/sim" "+firmware=${WORKSPACE}/build/firmware.hex"
    ;;
  native)
    mkdir -p "${REPO_ROOT}/build/sim-out" "${REPO_ROOT}/build"
    MAIN_SRC="${REPO_ROOT}/templates/hello-gpu/main.c"
    build_firmware "${REPO_ROOT}/build"
    build_verilator
    export DUMP_DIR="${REPO_ROOT}/build/sim-out"
    printf '{"cmd":"reset"}\n{"cmd":"run","cycles":5000000}\n' \
      | "${REPO_ROOT}/sim/obj_dir/Vtiny_gpu_top" "+firmware=${REPO_ROOT}/build/firmware.hex" \
      > "${REPO_ROOT}/build/sim-out/last-snapshot.json"
    check_pass_repo
    ;;
  *)
    echo "unknown mode: $MODE" >&2; exit 1 ;;
esac
