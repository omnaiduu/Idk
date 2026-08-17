`default_nettype none

module pwm (
    input  wire        clk,
    input  wire        resetn,
    input  wire [ 3:0] sel,
    input  wire        wr,
    input  wire [31:0] wdata,
    input  wire [ 3:0] wstrb,
    output reg  [31:0] rdata,
    output reg         ready,
    output wire [ 7:0] pwm0,
    output wire [ 7:0] pwm1,
    output wire [ 7:0] pwm2,
    output wire [ 7:0] pwm3,
    output wire [ 7:0] duty0,
    output wire [ 7:0] duty1,
    output wire [ 7:0] duty2,
    output wire [ 7:0] duty3
);
    reg [7:0] duty_reg [0:3];
    reg [7:0] counter;

    assign duty0 = duty_reg[0];
    assign duty1 = duty_reg[1];
    assign duty2 = duty_reg[2];
    assign duty3 = duty_reg[3];

    assign pwm0 = (counter < duty_reg[0]) ? 8'd255 : 8'd0;
    assign pwm1 = (counter < duty_reg[1]) ? 8'd255 : 8'd0;
    assign pwm2 = (counter < duty_reg[2]) ? 8'd255 : 8'd0;
    assign pwm3 = (counter < duty_reg[3]) ? 8'd255 : 8'd0;

    integer ch;

    always @(posedge clk) begin
        ready <= 1'b0;
        if (!resetn) begin
            counter <= 8'd0;
            for (ch = 0; ch < 4; ch = ch + 1)
                duty_reg[ch] <= 8'd0;
            rdata <= 32'h0;
        end else begin
            counter <= counter + 8'd1;
            if (|sel) begin
                ready <= 1'b1;
                for (ch = 0; ch < 4; ch = ch + 1) begin
                    if (sel[ch]) begin
                        if (wr && wstrb[0])
                            duty_reg[ch] <= wdata[7:0];
                        else if (!wr)
                            rdata <= {24'h0, duty_reg[ch]};
                    end
                end
            end
        end
    end
endmodule
