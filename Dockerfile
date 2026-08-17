# syntax=docker/dockerfile:1

# Stage 1: build React SPA
FROM node:22-bookworm AS web-build
RUN corepack enable && corepack prepare pnpm@latest --activate
WORKDIR /src/apps/web
COPY apps/web/package.json apps/web/pnpm-lock.yaml ./
RUN pnpm install --frozen-lockfile
COPY apps/web/ ./
RUN pnpm build

# Stage 2: build Go server
FROM golang:1.23-bookworm AS go-build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY cmd/ ./cmd/
COPY internal/ ./internal/
RUN CGO_ENABLED=0 go build -o /tiny-gpu-bench ./cmd/tiny-gpu-bench

# Stage 3: runtime with chip tools
FROM ubuntu:24.04 AS runtime
ENV DEBIAN_FRONTEND=noninteractive
RUN apt-get update && apt-get install -y --no-install-recommends \
    verilator \
    clang \
    lld \
    yosys \
    yosys-abc \
    nextpnr-ice40 \
    nextpnr-ice40-chipdb \
    fpga-icestorm \
    gcc-riscv64-unknown-elf \
    binutils-riscv64-unknown-elf \
    make \
    g++ \
    python3 \
    ca-certificates \
    && rm -rf /var/lib/apt/lists/*

WORKDIR /app
COPY --from=go-build /tiny-gpu-bench /app/tiny-gpu-bench
COPY --from=web-build /src/apps/web/dist /app/web
COPY hdl/ /app/hdl/
COPY firmware/ /app/firmware/
COPY sim/ /app/sim/
COPY scripts/ /app/scripts/
COPY templates/ /app/templates/

RUN chmod +x /app/scripts/dev-sim.sh

ENV HOST=0.0.0.0 \
    PORT=8741 \
    WEB_ROOT=/app/web \
    WORKSPACE=/data/workspace \
    SIM_TIMEOUT_SEC=120

EXPOSE 8741
VOLUME ["/data/workspace"]
ENTRYPOINT ["/app/tiny-gpu-bench"]
