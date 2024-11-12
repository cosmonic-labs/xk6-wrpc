FROM golang:1.23 AS build
WORKDIR /go/src/xk6-wrpc
RUN go install go.k6.io/xk6/cmd/xk6@latest
COPY . .
RUN xk6 build --with xk6-wrpc=.

FROM alpine:latest AS base
RUN apk add --no-cache ca-certificates && \
  addgroup -S app && adduser -S -G app app
COPY --from=build /go/src/xk6-wrpc/k6 /k6
USER app
ENTRYPOINT ["/k6"]
