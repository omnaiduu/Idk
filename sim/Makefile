# Tiny GPU Bench Verilator simulation build

ROOT       := $(abspath ..)
BUILD      := $(ROOT)/build
SIM_OUT    := $(BUILD)/sim-out
VERILATOR  := verilator

FW_ELF     := $(BUILD)/firmware.elf
FW_BIN     := $(BUILD)/firmware.bin
FW_HEX     := $(BUILD)/firmware.hex

GCC        := riscv64-unknown-elf-gcc
OBJCOPY    := riscv64-unknown-elf-objcopy

GCCFLAGS   := -march=rv32i -mabi=ilp32 -nostdlib -nostartfiles \
              -ffreestanding -fno-builtin -Os \
              -T $(ROOT)/firmware/link.ld -I $(ROOT)/firmware

MAIN_SRC   ?= $(ROOT)/templates/hello-gpu/main.c

SRC_V      := $(ROOT)/hdl/tiny_gpu_top.v \
              $(ROOT)/hdl/soc.v \
              $(ROOT)/hdl/picorv32.v \
              $(ROOT)/hdl/ram.v \
              $(ROOT)/hdl/gpio_led.v \
              $(ROOT)/hdl/btn.v \
              $(ROOT)/hdl/uart_tx.v \
              $(ROOT)/hdl/pwm.v \
              $(ROOT)/hdl/gpu.v

SRC_CPP    := main.cpp dpi_dump.cpp

.PHONY: all clean firmware sim run-test

all: sim

firmware: $(FW_HEX)

$(BUILD):
	mkdir -p $(BUILD) $(SIM_OUT)

$(FW_ELF): $(BUILD) $(ROOT)/firmware/start.S $(MAIN_SRC) \
           $(ROOT)/firmware/link.ld $(ROOT)/firmware/tgb.h
	$(GCC) $(GCCFLAGS) $(ROOT)/firmware/start.S $(MAIN_SRC) -o $@

$(FW_BIN): $(FW_ELF)
	$(OBJCOPY) -O binary $< $@

$(FW_HEX): $(FW_ELF)
	$(OBJCOPY) -O verilog $< $@

obj_dir/Vtiny_gpu_top: $(SRC_V) $(SRC_CPP) $(FW_HEX) | $(BUILD)
	$(VERILATOR) --cc --exe --build -j 1 \
		-Wno-fatal -Wno-WIDTH -Wno-UNOPTFLAT -Wno-DECLFILENAME \
		--top-module tiny_gpu_top \
		-Mdir obj_dir \
		-CFLAGS "-std=c++17 -I." \
		$(SRC_V) $(SRC_CPP)
	@touch $@

sim: obj_dir/Vtiny_gpu_top firmware

run-test: sim
	cd $(ROOT) && printf '{"cmd":"reset"}\n{"cmd":"run","cycles":5000000}\n' \
		| sim/obj_dir/Vtiny_gpu_top +firmware=$(FW_HEX)

clean:
	rm -rf obj_dir $(BUILD)/firmware.elf $(BUILD)/firmware.bin $(BUILD)/firmware.hex
