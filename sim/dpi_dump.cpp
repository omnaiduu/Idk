#include <cstdint>
#include <cstdio>
#include <cstring>
#include <fstream>
#include <string>

static std::string g_dump_dir = "build/sim-out";
static std::string g_uart_log;

// PLAN.md: last ≤1024 samples (ring, not first-1024).
static const size_t WAVE_CAP = 1024;
static uint8_t g_wave_clk[WAVE_CAP];
static uint8_t g_wave_gpu_busy[WAVE_CAP];
static uint8_t g_wave_led0[WAVE_CAP];
static uint8_t g_wave_pwm0[WAVE_CAP];
static size_t g_wave_n = 0;
static size_t g_wave_head = 0;

extern "C" {

void dpi_set_dump_dir(const char *dir) {
    if (dir && dir[0]) g_dump_dir = dir;
}

void dpi_reset_dumps() {
    g_uart_log.clear();
    g_wave_n = 0;
    g_wave_head = 0;
}

void dpi_uart_byte(unsigned char ch) {
    g_uart_log.push_back((char)ch);
}

void dpi_wave_sample(int clk, int gpu_busy, int led0, int pwm0) {
    g_wave_clk[g_wave_head] = (uint8_t)(clk & 1);
    g_wave_gpu_busy[g_wave_head] = (uint8_t)(gpu_busy & 1);
    g_wave_led0[g_wave_head] = (uint8_t)(led0 & 1);
    g_wave_pwm0[g_wave_head] = (uint8_t)(pwm0 & 0xff);
    g_wave_head = (g_wave_head + 1) % WAVE_CAP;
    if (g_wave_n < WAVE_CAP) g_wave_n++;
}

void dpi_dump_uart(const char *path, const char *text) {
    std::ofstream f(path, std::ios::binary);
    f.write(text, std::strlen(text));
}

void dpi_dump_leds(const char *path, uint32_t leds) {
    std::ofstream f(path);
    f << "{\"leds\":" << leds << "}\n";
}

void dpi_dump_pwm(const char *path, uint32_t d0, uint32_t d1, uint32_t d2, uint32_t d3) {
    std::ofstream f(path);
    f << "{\"duty\":[" << d0 << "," << d1 << "," << d2 << "," << d3 << "]}\n";
}

void dpi_dump_fb(const char *path, const uint8_t *fb, int size) {
    std::ofstream f(path, std::ios::binary);
    f.write(reinterpret_cast<const char *>(fb), size);
}

static void dump_ring(std::ofstream &f, const uint8_t *buf) {
    size_t start = 0;
    if (g_wave_n == WAVE_CAP) start = g_wave_head;
    for (size_t i = 0; i < g_wave_n; i++) {
        if (i) f << ',';
        f << (unsigned)buf[(start + i) % WAVE_CAP];
    }
}

void dpi_dump_waves(const char *path) {
    std::ofstream f(path);
    f << "{\"clk\":[";
    dump_ring(f, g_wave_clk);
    f << "],\"gpu_busy\":[";
    dump_ring(f, g_wave_gpu_busy);
    f << "],\"led0\":[";
    dump_ring(f, g_wave_led0);
    f << "],\"pwm0\":[";
    dump_ring(f, g_wave_pwm0);
    f << "]}\n";
}

void dpi_flush_dumps(int leds, int d0, int d1, int d2, int d3,
                     const uint8_t *fb, int fb_size) {
    std::string uart_path  = g_dump_dir + "/uart.log";
    std::string leds_path  = g_dump_dir + "/leds.json";
    std::string pwm_path   = g_dump_dir + "/pwm.json";
    std::string fb_path    = g_dump_dir + "/fb.bin";
    std::string waves_path = g_dump_dir + "/waves.json";

    dpi_dump_uart(uart_path.c_str(), g_uart_log.c_str());
    dpi_dump_leds(leds_path.c_str(), (uint32_t)leds);
    dpi_dump_pwm(pwm_path.c_str(), d0, d1, d2, d3);
    dpi_dump_fb(fb_path.c_str(), fb, fb_size);
    dpi_dump_waves(waves_path.c_str());
}

}  // extern "C"
