`default_nettype none

module ram #(
    parameter ADDR_WIDTH = 14
) (
    input  wire        clk,
    input  wire        resetn,
    input  wire        sel,
    input  wire        wr,
    input  wire [31:0] addr,
    input  wire [31:0] wdata,
    input  wire [ 3:0] wstrb,
    output reg  [31:0] rdata,
    output reg         ready
);
    localparam DEPTH = 1 << ADDR_WIDTH;
    reg [7:0] mem [0:DEPTH-1];

    wire [ADDR_WIDTH-1:0] byte_addr = addr[ADDR_WIDTH-1:0];

    integer i;
    reg [1023:0] firmware_path;

    initial begin
        for (i = 0; i < DEPTH; i = i + 1)
            mem[i] = 8'h00;
        if ($value$plusargs("firmware=%s", firmware_path))
            $readmemh(firmware_path, mem);
    end

    always @(posedge clk) begin
        ready <= 1'b0;
        if (!resetn) begin
            rdata <= 32'h0;
        end else if (sel) begin
            ready <= 1'b1;
            if (wr) begin
                if (wstrb[0]) mem[byte_addr + 0] <= wdata[7:0];
                if (wstrb[1]) mem[byte_addr + 1] <= wdata[15:8];
                if (wstrb[2]) mem[byte_addr + 2] <= wdata[23:16];
                if (wstrb[3]) mem[byte_addr + 3] <= wdata[31:24];
            end else begin
                rdata <= {
                    mem[byte_addr + 3],
                    mem[byte_addr + 2],
                    mem[byte_addr + 1],
                    mem[byte_addr + 0]
                };
            end
        end
    end
endmodule
