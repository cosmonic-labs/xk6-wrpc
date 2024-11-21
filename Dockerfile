FROM ghcr.io/bytecodealliance/wrpc:cb43ec8 AS wrpc

FROM golang:1.23 AS build
WORKDIR /go/src/xk6-wrpc
RUN go install go.k6.io/xk6/cmd/xk6@latest
RUN CGO_ENABLED=0 go install github.com/nats-io/natscli/nats@latest
RUN CGO_ENABLED=0 go install github.com/rakyll/hey@v0.1.4
COPY . .
RUN xk6 build --with xk6-wrpc=. --with github.com/grafana/xk6-dashboard@latest


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
COPY --from=build /go/src/xk6-wrpc/blaster-component/blaster-component.wasm /blaster-component.wasm
USER k6
ENTRYPOINT ["/usr/bin/k6"]
