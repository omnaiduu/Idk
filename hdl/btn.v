`default_nettype none

// Button input register at 0x10000004 (read-only, bit0)
module btn (
    input  wire        clk,
    input  wire        resetn,
    input  wire        sel,
    input  wire        wr,
    input  wire        btn_in,
    output reg  [31:0] rdata,
    output reg         ready
);
    always @(posedge clk) begin
        ready <= 1'b0;
        if (!resetn) begin
            rdata <= 32'h0;
        end else if (sel) begin
            ready <= 1'b1;
            if (!wr)
                rdata <= {31'h0, btn_in};
        end
    end
endmodule

`default_nettype wire
