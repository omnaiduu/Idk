#ifndef TGB_H
#define TGB_H

#include <stdint.h>

#define REG32(addr) (*(volatile uint32_t *)(addr))

#define LEDS      REG32(0x10000000u)
#define BTN       REG32(0x10000004u)
#define UART_TX   REG32(0x20000000u)
#define UART_STAT REG32(0x20000004u)
#define PWM0      REG32(0x30000000u)
#define PWM1      REG32(0x30000004u)
#define PWM2      REG32(0x30000008u)
#define PWM3      REG32(0x3000000Cu)
#define GPU_CMD   REG32(0x40000000u)
#define GPU_STAT  REG32(0x40000004u)

#define UART_TX_READY  1u
#define GPU_BUSY       1u

#define GPU_OP_CLEAR     0x10000000u
#define GPU_OP_SET_COLOR 0x20000000u
#define GPU_OP_PUT_PIXEL 0x30000000u
#define GPU_OP_FILL_RECT 0x40000000u

static inline void uart_putc(char c) {
    while ((UART_STAT & UART_TX_READY) == 0) { }
    UART_TX = (uint32_t)(unsigned char)c;
}

static inline void uart_puts(const char *s) {
    while (*s) uart_putc(*s++);
}

static inline void gpu_wait(void) {
    while (GPU_STAT & GPU_BUSY) { }
}

static inline void gpu_clear(uint8_t color) {
    gpu_wait();
    GPU_CMD = GPU_OP_CLEAR | color;
}

static inline void gpu_set_color(uint8_t color) {
    gpu_wait();
    GPU_CMD = GPU_OP_SET_COLOR | color;
}

static inline void gpu_put_pixel(uint8_t x, uint8_t y) {
    gpu_wait();
    GPU_CMD = GPU_OP_PUT_PIXEL | ((uint32_t)x << 8) | y;
}

static inline void gpu_fill_rect(uint8_t x, uint8_t y, uint8_t w, uint8_t h) {
    gpu_wait();
    GPU_CMD = GPU_OP_FILL_RECT
            | ((uint32_t)x << 22)
            | ((uint32_t)y << 16)
            | ((uint32_t)w << 8)
            | h;
}

static inline void delay(volatile uint32_t n) {
    while (n--) { }
}

#endif
