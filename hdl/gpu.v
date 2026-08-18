`default_nettype none

// Tiny GPU: 64x64 RGB332 framebuffer, 4 pixels/cycle
module gpu (
    input  wire        clk,
    input  wire        resetn,
    input  wire        sel_cmd,
    input  wire        sel_stat,
    input  wire        wr,
    input  wire [31:0] wdata,
    output reg  [31:0] rdata_stat,
    output reg         ready_cmd,
    output reg         ready_stat,
    output wire        busy
);
    localparam FB_W    = 64;
    localparam FB_H    = 64;
    localparam FB_SIZE = 4096;

    localparam OP_CLEAR     = 4'h1;
    localparam OP_SET_COLOR = 4'h2;
    localparam OP_PUT_PIXEL = 4'h3;
    localparam OP_FILL_RECT = 4'h4;

    localparam ST_IDLE  = 2'd0;
    localparam ST_CLEAR = 2'd1;
    localparam ST_FILL  = 2'd2;

    reg [7:0] framebuffer [0:FB_SIZE-1] /* verilator public_flat_rd */;
    reg [7:0] cur_color;
    reg       busy_reg;
    reg [1:0] state;

    reg [11:0] pixel_idx;
    reg [11:0] pixel_end;

    reg [5:0] rect_x, rect_y, rect_w, rect_h;
    reg [5:0] fill_px, fill_py;
    reg [11:0] fill_off;
    reg [11:0] idx0, idx1, idx2, idx3;
    reg [5:0]  px0, py0, px1, py1, px2, py2, px3, py3;

    reg [31:0] cmd_latch;
    reg        cmd_pending;

    assign busy = busy_reg;

    integer i, n;

    function [11:0] f_idx;
        input [5:0] x;
        input [5:0] y;
        begin
            f_idx = {y, 6'b0} + x;
        end
    endfunction

    initial begin
        for (i = 0; i < FB_SIZE; i = i + 1)
            framebuffer[i] = 8'h00;
    end

    always @(posedge clk) begin
        ready_cmd  <= 1'b0;
        ready_stat <= 1'b0;

        if (!resetn) begin
            busy_reg    <= 1'b0;
            cur_color   <= 8'h00;
            state       <= ST_IDLE;
            pixel_idx   <= 12'd0;
            pixel_end   <= 12'd0;
            cmd_pending <= 1'b0;
            cmd_latch   <= 32'h0;
            rdata_stat  <= 32'h0;
        end else begin
            if (sel_stat) begin
                ready_stat <= 1'b1;
                if (!wr)
                    rdata_stat <= {31'h0, busy_reg};
            end

            if (sel_cmd) begin
                ready_cmd <= 1'b1;
                if (wr && !cmd_pending && !busy_reg) begin
                    cmd_latch   <= wdata;
                    cmd_pending <= 1'b1;
                end
            end

            if (cmd_pending && !busy_reg) begin
                cmd_pending <= 1'b0;
                case (cmd_latch[31:28])
                    OP_CLEAR: begin
                        cur_color <= cmd_latch[7:0];
                        pixel_idx <= 12'd0;
                        pixel_end <= 12'd4096;
                        state     <= ST_CLEAR;
                        busy_reg  <= 1'b1;
                    end
                    OP_SET_COLOR: begin
                        cur_color <= cmd_latch[7:0];
                    end
                    OP_PUT_PIXEL: begin
                        framebuffer[f_idx(cmd_latch[13:8], cmd_latch[5:0])] <= cur_color;
                    end
                    OP_FILL_RECT: begin
                        rect_x    <= cmd_latch[27:22];
                        rect_y    <= cmd_latch[21:16];
                        rect_w    <= cmd_latch[13:8];
                        rect_h    <= cmd_latch[5:0];
                        pixel_idx <= 12'd0;
                        pixel_end <= cmd_latch[13:8] * cmd_latch[5:0];
                        state     <= ST_FILL;
                        busy_reg  <= 1'b1;
                    end
                    default: ;
                endcase
            end

            if (busy_reg) begin
                case (state)
                    ST_CLEAR: begin
                        idx0 = pixel_idx + 12'd0;
                        idx1 = pixel_idx + 12'd1;
                        idx2 = pixel_idx + 12'd2;
                        idx3 = pixel_idx + 12'd3;
                        if (idx0 < pixel_end) framebuffer[idx0] <= cur_color;
                        if (idx1 < pixel_end) framebuffer[idx1] <= cur_color;
                        if (idx2 < pixel_end) framebuffer[idx2] <= cur_color;
                        if (idx3 < pixel_end) framebuffer[idx3] <= cur_color;
                        if (pixel_idx + 12'd4 >= pixel_end) begin
                            busy_reg  <= 1'b0;
                            state     <= ST_IDLE;
                            pixel_idx <= 12'd0;
                        end else begin
                            pixel_idx <= pixel_idx + 12'd4;
                        end
                    end
                    ST_FILL: begin
                        idx0 = pixel_idx + 12'd0;
                        idx1 = pixel_idx + 12'd1;
                        idx2 = pixel_idx + 12'd2;
                        idx3 = pixel_idx + 12'd3;
                        if (idx0 < pixel_end) begin
                            px0 = rect_x + idx0 % rect_w;
                            py0 = rect_y + idx0 / rect_w;
                            if (px0 < FB_W && py0 < FB_H)
                                framebuffer[f_idx(px0, py0)] <= cur_color;
                        end
                        if (idx1 < pixel_end) begin
                            px1 = rect_x + idx1 % rect_w;
                            py1 = rect_y + idx1 / rect_w;
                            if (px1 < FB_W && py1 < FB_H)
                                framebuffer[f_idx(px1, py1)] <= cur_color;
                        end
                        if (idx2 < pixel_end) begin
                            px2 = rect_x + idx2 % rect_w;
                            py2 = rect_y + idx2 / rect_w;
                            if (px2 < FB_W && py2 < FB_H)
                                framebuffer[f_idx(px2, py2)] <= cur_color;
                        end
                        if (idx3 < pixel_end) begin
                            px3 = rect_x + idx3 % rect_w;
                            py3 = rect_y + idx3 / rect_w;
                            if (px3 < FB_W && py3 < FB_H)
                                framebuffer[f_idx(px3, py3)] <= cur_color;
                        end
                        if (pixel_idx + 12'd4 >= pixel_end) begin
                            busy_reg  <= 1'b0;
                            state     <= ST_IDLE;
                            pixel_idx <= 12'd0;
                        end else begin
                            pixel_idx <= pixel_idx + 12'd4;
                        end
                    end
                    default: ;
                endcase
            end
        end
    end
endmodule
