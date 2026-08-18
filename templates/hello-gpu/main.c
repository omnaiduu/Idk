#include "tgb.h"

// 5x7 bitmap glyphs (bit0 = left column)
static const uint8_t glyph_O[7] = {
    0x0E, 0x11, 0x11, 0x11, 0x11, 0x11, 0x0E
};

static const uint8_t glyph_M[7] = {
    0x11, 0x1B, 0x15, 0x11, 0x11, 0x11, 0x11
};

static void draw_glyph(uint8_t x0, uint8_t y0, const uint8_t *rows, uint8_t color) {
    gpu_set_color(color);
    for (uint8_t y = 0; y < 7; y++) {
        uint8_t row = rows[y];
        for (uint8_t x = 0; x < 5; x++) {
            if (row & (1u << (4 - x)))
                gpu_put_pixel(x0 + x, y0 + y);
        }
    }
}

static void led_chaser(void) {
    for (uint8_t pass = 0; pass < 2; pass++) {
        for (uint8_t i = 0; i < 8; i++) {
            LEDS = (uint32_t)(1u << i);
            delay(10000u);
        }
    }
    LEDS = 0xFFu;
}

int main(void) {
  // PWM duties: 32, 96, 160, 224
    PWM0 = 32u;
    PWM1 = 96u;
    PWM2 = 160u;
    PWM3 = 224u;

    uart_puts("hello from tiny-gpu\n");

    led_chaser();

    // GPU scene: dark background, colored rects, OM letters
    gpu_clear(0x02u);  // near-black blue background
    gpu_wait();

    gpu_set_color(0xE0u);  // red rect (top-left)
    gpu_fill_rect(2, 2, 20, 10);
    gpu_wait();

    gpu_set_color(0x1Cu);  // green rect (top-right)
    gpu_fill_rect(42, 2, 20, 10);
    gpu_wait();

    gpu_set_color(0xFCu);  // orange/amber rect (bottom)
    gpu_fill_rect(10, 50, 44, 10);
    gpu_wait();

    draw_glyph(18, 28, glyph_O, 0xFFu);
    draw_glyph(30, 28, glyph_M, 0xFFu);
    gpu_wait();

    LEDS = 0xAAu;
    return 0;
}
