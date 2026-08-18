`default_nettype none

module gpio_led (
    input  wire        clk,
    input  wire        resetn,
    input  wire        sel,
    input  wire        wr,
    input  wire [31:0] wdata,
    input  wire [ 3:0] wstrb,
    output reg  [31:0] rdata,
    output reg         ready,
    output wire [ 7:0] leds
);
    reg [7:0] led_reg;
    assign leds = led_reg;

    always @(posedge clk) begin
        ready <= 1'b0;
        if (!resetn) begin
            led_reg <= 8'h00;
            rdata   <= 32'h0;
        end else if (sel) begin
            ready <= 1'b1;
            if (wr && wstrb[0])
                led_reg <= wdata[7:0];
            else if (!wr)
                rdata <= {24'h0, led_reg};
        end
    end
endmodule
