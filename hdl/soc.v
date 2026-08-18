// Tiny GPU Bench SoC — PicoRV32 + peripherals
module soc (
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
    // PicoRV32 memory bus
    wire        mem_valid;
    wire        mem_instr;
    wire        mem_ready;
    wire [31:0] mem_addr;
    wire [31:0] mem_wdata;
    wire [ 3:0] mem_wstrb;
    wire [31:0] mem_rdata;

    wire mem_wr = mem_valid && |mem_wstrb;

    wire        pcpi_valid;
    wire [31:0] pcpi_insn;
    wire [31:0] pcpi_rs1;
    wire [31:0] pcpi_rs2;
    wire        pcpi_wait;

    // Address decode
    wire sel_ram      = mem_valid && (mem_addr[31:28] == 4'h0);
    wire sel_gpio     = mem_valid && (mem_addr[31:4] == 28'h1000000);
    wire sel_led      = sel_gpio && (mem_addr[3:2] == 2'b00);
    wire sel_btn      = sel_gpio && (mem_addr[3:2] == 2'b01);
    wire sel_uart     = mem_valid && (mem_addr[31:4] == 28'h2000000);
    wire sel_uart_tx  = sel_uart && (mem_addr[3:2] == 2'b00);
    wire sel_uart_st  = sel_uart && (mem_addr[3:2] == 2'b01);
    wire sel_pwm      = mem_valid && (mem_addr[31:4] == 28'h3000000);
    wire sel_pwm_ch0  = sel_pwm && (mem_addr[3:2] == 2'b00);
    wire sel_pwm_ch1  = sel_pwm && (mem_addr[3:2] == 2'b01);
    wire sel_pwm_ch2  = sel_pwm && (mem_addr[3:2] == 2'b10);
    wire sel_pwm_ch3  = sel_pwm && (mem_addr[3:2] == 2'b11);
    wire sel_gpu      = mem_valid && (mem_addr[31:4] == 28'h4000000);
    wire sel_gpu_cmd  = sel_gpu && (mem_addr[3:2] == 2'b00);
    wire sel_gpu_stat = sel_gpu && (mem_addr[3:2] == 2'b01);

    wire [3:0] pwm_sel = {sel_pwm_ch3, sel_pwm_ch2, sel_pwm_ch1, sel_pwm_ch0};

    // Peripheral read data and ready
    wire [31:0] ram_rdata;
    wire        ram_ready;
    wire [31:0] led_rdata;
    wire        led_ready;
    wire [31:0] btn_rdata;
    wire        btn_ready;
    wire [31:0] uart_rdata_data;
    wire [31:0] uart_rdata_stat;
    wire        uart_ready_data;
    wire        uart_ready_stat;
    wire [31:0] pwm_rdata;
    wire        pwm_ready;
    wire [31:0] gpu_rdata_stat;
    wire        gpu_ready_cmd;
    wire        gpu_ready_stat;

    assign mem_rdata = ram_ready     ? ram_rdata       :
                       led_ready      ? led_rdata       :
                       btn_ready      ? btn_rdata       :
                       uart_ready_data ? uart_rdata_data :
                       uart_ready_stat ? uart_rdata_stat :
                       pwm_ready      ? pwm_rdata       :
                       gpu_ready_stat ? gpu_rdata_stat  :
                       32'h0;

    assign mem_ready = (sel_ram      && ram_ready)      ||
                       (sel_led      && led_ready)      ||
                       (sel_btn      && btn_ready)      ||
                       (sel_uart_tx  && uart_ready_data)||
                       (sel_uart_st  && uart_ready_stat)||
                       (sel_pwm      && pwm_ready)      ||
                       (sel_gpu_cmd  && gpu_ready_cmd)  ||
                       (sel_gpu_stat && gpu_ready_stat) ||
                       (!mem_valid);

    picorv32 #(
        .ENABLE_COUNTERS     (1),
        .ENABLE_COUNTERS64   (1),
        .ENABLE_REGS_16_31   (1),
        .ENABLE_REGS_DUALPORT(1),
        .TWO_STAGE_SHIFT     (1),
        .BARREL_SHIFTER      (0),
        .TWO_CYCLE_COMPARE   (0),
        .TWO_CYCLE_ALU       (0),
        .COMPRESSED_ISA      (0),
        .CATCH_MISALIGN      (1),
        .CATCH_ILLINSN       (1),
        .ENABLE_PCPI         (0),
        .ENABLE_MUL          (0),
        .ENABLE_FAST_MUL     (0),
        .ENABLE_DIV          (0),
        .ENABLE_IRQ          (0),
        .ENABLE_IRQ_QREGS    (0),
        .ENABLE_IRQ_TIMER    (0),
        .ENABLE_TRACE        (0),
        .REGS_INIT_ZERO      (1),
        .MASKED_IRQ          (32'h0),
        .LATCHED_IRQ         (32'h0),
        .PROGADDR_RESET      (32'h00000000),
        .PROGADDR_IRQ        (32'h00000010)
    ) cpu (
        .clk         (clk),
        .resetn      (resetn),
        .trap        (trap),
        .mem_valid   (mem_valid),
        .mem_instr   (mem_instr),
        .mem_ready   (mem_ready),
        .mem_addr    (mem_addr),
        .mem_wdata   (mem_wdata),
        .mem_wstrb   (mem_wstrb),
        .mem_rdata   (mem_rdata),
        .mem_la_read (),
        .mem_la_write(),
        .mem_la_addr (),
        .mem_la_wdata(),
        .mem_la_wstrb(),
        .pcpi_valid  (pcpi_valid),
        .pcpi_insn   (pcpi_insn),
        .pcpi_rs1    (pcpi_rs1),
        .pcpi_rs2    (pcpi_rs2),
        .pcpi_wr     (1'b0),
        .pcpi_rd     (32'h0),
        .pcpi_wait   (pcpi_wait),
        .pcpi_ready  (1'b0),
        .irq         (32'h0),
        .eoi         (),
        .trace_valid (),
        .trace_data  ()
    );

    ram #(
        .ADDR_WIDTH(14)
    ) u_ram (
        .clk    (clk),
        .resetn (resetn),
        .sel    (sel_ram),
        .wr     (mem_wr),
        .addr   (mem_addr),
        .wdata  (mem_wdata),
        .wstrb  (mem_wstrb),
        .rdata  (ram_rdata),
        .ready  (ram_ready)
    );

    gpio_led u_led (
        .clk    (clk),
        .resetn (resetn),
        .sel    (sel_led),
        .wr     (mem_wr),
        .wdata  (mem_wdata),
        .wstrb  (mem_wstrb),
        .rdata  (led_rdata),
        .ready  (led_ready),
        .leds   (leds)
    );

    btn u_btn (
        .clk    (clk),
        .resetn (resetn),
        .sel    (sel_btn),
        .wr     (mem_wr),
        .btn_in (btn_in),
        .rdata  (btn_rdata),
        .ready  (btn_ready)
    );

    uart_tx #(
        .CLK_HZ(10000000),
        .BAUD  (115200)
    ) u_uart (
        .clk          (clk),
        .resetn       (resetn),
        .sel_data     (sel_uart_tx),
        .sel_stat     (sel_uart_st),
        .wr           (mem_wr),
        .wdata        (mem_wdata),
        .rdata_data   (uart_rdata_data),
        .rdata_stat   (uart_rdata_stat),
        .ready_data   (uart_ready_data),
        .ready_stat   (uart_ready_stat),
        .tx           (uart_tx),
        .tx_byte_valid(uart_byte_valid),
        .tx_byte      (uart_byte)
    );

    pwm u_pwm (
        .clk      (clk),
        .resetn   (resetn),
        .sel      (pwm_sel),
        .wr       (mem_wr),
        .wdata    (mem_wdata),
        .wstrb    (mem_wstrb),
        .rdata    (pwm_rdata),
        .ready    (pwm_ready),
        .pwm0     (pwm0_out),
        .pwm1     (pwm1_out),
        .pwm2     (pwm2_out),
        .pwm3     (pwm3_out),
        .duty0    (duty0),
        .duty1    (duty1),
        .duty2    (duty2),
        .duty3    (duty3)
    );

    gpu u_gpu (
        .clk        (clk),
        .resetn     (resetn),
        .sel_cmd    (sel_gpu_cmd),
        .sel_stat   (sel_gpu_stat),
        .wr         (mem_wr),
        .wdata      (mem_wdata),
        .rdata_stat (gpu_rdata_stat),
        .ready_cmd  (gpu_ready_cmd),
        .ready_stat (gpu_ready_stat),
        .busy       (gpu_busy)
    );
endmodule
