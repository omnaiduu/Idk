`default_nettype none

module tiny_gpu_top (
    input  wire        clk,
    input  wire        resetn,
    input  wire        btn_in,
    output wire [ 7:0] leds,
    output wire [ 7:0] pwm0_out,
    output wire [ 7:0] pwm1_out,
    output wire [ 7:0] pwm2_out,
    output wire [ 7:0] pwm3_out,
    output wire [ 7:0] duty0,
    output wire [ 7:0] duty1,
    output wire [ 7:0] duty2,
    output wire [ 7:0] duty3,
    output wire        uart_tx,
    output wire        uart_byte_valid,
    output wire [ 7:0] uart_byte,
    output wire        gpu_busy,
    output wire        trap
);
    soc u_soc (
        .clk          (clk),
        .resetn       (resetn),
        .btn_in       (btn_in),
        .leds         (leds),
        .pwm0_out     (pwm0_out),
        .pwm1_out     (pwm1_out),
        .pwm2_out     (pwm2_out),
        .pwm3_out     (pwm3_out),
        .duty0        (duty0),
        .duty1        (duty1),
        .duty2        (duty2),
        .duty3        (duty3),
        .uart_tx      (uart_tx),
        .uart_byte_valid(uart_byte_valid),
        .uart_byte    (uart_byte),
        .gpu_busy     (gpu_busy),
        .trap         (trap)
    );
endmodule

`default_nettype wire
