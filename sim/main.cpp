#include <verilated.h>
#include "Vtiny_gpu_top.h"
#include "Vtiny_gpu_top___024root.h"

#include <cstdint>
#include <cstdio>
#include <cstdlib>
#include <cstring>
#include <iostream>
#include <string>
#include <sys/stat.h>

extern "C" {
void dpi_set_dump_dir(const char *dir);
void dpi_reset_dumps();
void dpi_uart_byte(unsigned char ch);
void dpi_wave_sample(int clk, int gpu_busy, int led0, int pwm0);
void dpi_flush_dumps(int leds, int d0, int d1, int d2, int d3,
                     const uint8_t *fb, int fb_size);
}

static const char *OUT_DIR = "build/sim-out";
static const int FB_SIZE = 4096;

static Vtiny_gpu_top *g_top = nullptr;
static vluint64_t g_cycles = 0;
static int g_btn = 0;

static void ensure_out_dir() {
    mkdir("build", 0755);
    mkdir(OUT_DIR, 0755);
}

static void tick(int n = 1) {
    for (int i = 0; i < n; i++) {
        g_top->clk = 0;
        g_top->eval();

        g_top->clk = 1;
        g_top->eval();
        if (g_top->uart_byte_valid)
            dpi_uart_byte(g_top->uart_byte);
        dpi_wave_sample(g_top->clk, g_top->gpu_busy, g_top->leds & 1, g_top->duty0);

        g_cycles++;
    }
}

static void write_dumps() {
    uint8_t fb[FB_SIZE];
    auto *root = g_top->rootp;
    for (int i = 0; i < FB_SIZE; i++)
        fb[i] = root->tiny_gpu_top__DOT__u_soc__DOT__u_gpu__DOT__framebuffer[i];

    dpi_flush_dumps(g_top->leds, g_top->duty0, g_top->duty1, g_top->duty2, g_top->duty3,
                    fb, FB_SIZE);
}

static void emit_snapshot(bool halted) {
    write_dumps();
    std::cout << "{"
              << "\"cycle\":" << g_cycles << ","
              << "\"halted\":" << (halted ? "true" : "false") << ","
              << "\"dumps\":{"
              << "\"uart\":\"" << OUT_DIR << "/uart.log\","
              << "\"leds\":\"" << OUT_DIR << "/leds.json\","
              << "\"pwm\":\"" << OUT_DIR << "/pwm.json\","
              << "\"fb\":\"" << OUT_DIR << "/fb.bin\","
              << "\"waves\":\"" << OUT_DIR << "/waves.json\""
              << "}"
              << "}" << std::endl;
}

static void do_reset() {
    dpi_reset_dumps();
    g_top->resetn = 0;
    g_top->btn_in = g_btn;
    tick(10);
    g_top->resetn = 1;
    tick(2);
    g_cycles = 0;
}

static bool parse_cmd(const std::string &line, std::string &cmd, int &value) {
    if (line.find("reset") != std::string::npos) { cmd = "reset"; return true; }
    if (line.find("quit") != std::string::npos)  { cmd = "quit";  return true; }
    if (line.find("step") != std::string::npos)  { cmd = "step";  return true; }
    if (line.find("run") != std::string::npos) {
        cmd = "run";
        value = 2000000;
        auto pos = line.find("cycles");
        if (pos != std::string::npos) {
            const char *p = line.c_str() + pos;
            while (*p && (*p < '0' || *p > '9')) p++;
            if (*p) value = std::atoi(p);
        }
        return true;
    }
    if (line.find("set_btn") != std::string::npos) {
        cmd = "set_btn";
        value = (line.find("true") != std::string::npos || line.find(":1") != std::string::npos) ? 1 : 0;
        return true;
    }
    return false;
}

int main(int argc, char **argv) {
    Verilated::commandArgs(argc, argv);
    const char *dump = getenv("DUMP_DIR");
    if (dump) dpi_set_dump_dir(dump);

    ensure_out_dir();
    g_top = new Vtiny_gpu_top;
    g_top->clk = 0;
    g_top->resetn = 0;
    g_top->btn_in = 0;
    g_top->eval();

    do_reset();
    emit_snapshot(false);

    std::string line;
    while (std::getline(std::cin, line)) {
        if (line.empty()) continue;

        std::string cmd;
        int value = 0;
        if (!parse_cmd(line, cmd, value)) {
            std::cerr << "{\"error\":\"unknown command\"}" << std::endl;
            continue;
        }

        if (cmd == "quit") {
            emit_snapshot(g_top->trap);
            break;
        }
        if (cmd == "reset") {
            do_reset();
            emit_snapshot(false);
            continue;
        }
        if (cmd == "set_btn") {
            g_btn = value;
            g_top->btn_in = g_btn;
            emit_snapshot(g_top->trap);
            continue;
        }
        if (cmd == "step") {
            tick(1);
            emit_snapshot(g_top->trap);
            continue;
        }
        if (cmd == "run") {
            bool halted = false;
            int max_cycles = value > 0 ? value : 2000000;
            for (int i = 0; i < max_cycles; i++) {
                tick(1);
                if (g_top->trap) {
                    halted = true;
                    break;
                }
            }
            emit_snapshot(halted || g_top->trap);
        }
    }

    delete g_top;
    return 0;
}
