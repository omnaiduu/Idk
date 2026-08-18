`default_nettype none

module uart_tx #(
    parameter CLK_HZ = 10000000,
    parameter BAUD   = 115200
) (
    input  wire        clk,
    input  wire        resetn,
    input  wire        sel_data,
    input  wire        sel_stat,
    input  wire        wr,
    input  wire [31:0] wdata,
    output reg  [31:0] rdata_data,
    output reg  [31:0] rdata_stat,
    output reg         ready_data,
    output reg         ready_stat,
    output reg         tx,
    output reg         tx_byte_valid,
    output reg  [ 7:0] tx_byte
);
    localparam BAUD_DIV = CLK_HZ / BAUD;

    reg        busy;
    reg [15:0] baud_cnt;
    reg [3:0]  bit_cnt;
    reg [9:0]  shifter;
    reg [7:0]  tx_data_reg;

    wire tx_ready = !busy;

    always @(posedge clk) begin
        ready_data    <= 1'b0;
        ready_stat    <= 1'b0;
        tx_byte_valid <= 1'b0;

        if (!resetn) begin
            busy        <= 1'b0;
            baud_cnt    <= 16'd0;
            bit_cnt     <= 4'd0;
            shifter     <= 10'd0;
            tx          <= 1'b1;
            tx_data_reg <= 8'h00;
            rdata_data  <= 32'h0;
            rdata_stat  <= 32'h0;
        end else begin
            if (sel_stat) begin
                ready_stat <= 1'b1;
                if (!wr)
                    rdata_stat <= {31'h0, tx_ready};
            end

            if (sel_data) begin
                ready_data <= 1'b1;
                if (wr && tx_ready) begin
                    tx_data_reg <= wdata[7:0];
                    busy        <= 1'b1;
                    baud_cnt    <= 16'd0;
                    bit_cnt     <= 4'd0;
                    shifter     <= {1'b1, wdata[7:0], 1'b0};
                    tx          <= 1'b0;
                end
            end

            if (busy) begin
                if (baud_cnt >= BAUD_DIV - 1) begin
                    baud_cnt <= 16'd0;
                    tx       <= shifter[0];
                    shifter  <= {1'b1, shifter[9:1]};
                    bit_cnt  <= bit_cnt + 4'd1;
                    if (bit_cnt == 4'd9) begin
                        busy          <= 1'b0;
                        tx            <= 1'b1;
                        tx_byte       <= tx_data_reg;
                        tx_byte_valid <= 1'b1;
                    end
                end else begin
                    baud_cnt <= baud_cnt + 16'd1;
                end
            end
        end
    end
endmodule
