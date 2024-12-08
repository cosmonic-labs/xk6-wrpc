# syntax=docker/dockerfile:1

FROM ghcr.io/bytecodealliance/wrpc:cb43ec8 AS wrpc

FROM golang:1.23.4 AS build
ENV TINYGO_RELEASE=0.34.0
ENV WASMTOOLS_VERSION=1.221.2
ENV PATH=${PATH}:/usr/local/tinygo/bin
ARG TARGETOS
ARG TARGETARCH

RUN <<EOF
cd /tmp
wget -q https://github.com/tinygo-org/tinygo/releases/download/v${TINYGO_RELEASE}/tinygo${TINYGO_RELEASE}.${TARGETOS}-${TARGETARCH}.tar.gz
tar zxf tinygo${TINYGO_RELEASE}.${TARGETOS}-${TARGETARCH}.tar.gz -C /usr/local

wget -q https://github.com/bytecodealliance/wasm-tools/releases/download/v${WASMTOOLS_VERSION}/wasm-tools-${WASMTOOLS_VERSION}-x86_64-linux.tar.gz
tar zxf wasm-tools-${WASMTOOLS_VERSION}-x86_64-linux.tar.gz
cp wasm-tools-${WASMTOOLS_VERSION}-x86_64-linux/wasm-tools /usr/local/bin/
rm -f *gz
EOF

RUN <<EOF
go install go.k6.io/xk6/cmd/xk6@latest
CGO_ENABLED=0 go install github.com/nats-io/natscli/nats@v0.1.5
CGO_ENABLED=0 go install github.com/rakyll/hey@v0.1.4
EOF

WORKDIR /go/src/xk6-wrpc

COPY . .
RUN make build

FROM alpine:latest AS base
RUN apk add --no-cache \
  bash \
  ngrep \
  curl \
  jq \
  ca-certificates \
  && addgroup -S k6 \
  && adduser -S -s /bin/bash -G k6 k6
COPY --from=wrpc /bin/wrpc-wasmtime /usr/bin/wrpc-wasmtime
COPY --from=build /go/bin/nats /usr/bin/nats
COPY --from=build /go/bin/hey /usr/bin/hey
COPY --from=build /go/src/xk6-wrpc/k6 /usr/bin/k6
COPY --from=build /go/src/xk6-wrpc/components/blaster/blaster.wasm /components/
USER k6
ENTRYPOINT ["/usr/bin/k6"]
