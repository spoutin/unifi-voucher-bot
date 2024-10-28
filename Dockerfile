# syntax=docker/dockerfile:1

FROM golang:1.23 AS builder

COPY src/go.mod src/go.sum ./
RUN go mod download
COPY src/*.go ./
RUN CGO_ENABLED=0 GOOS=linux go build -o wifi-voucher-bot

FROM gcr.io/distroless/base-debian11 AS build-release-stage
WORKDIR /
COPY --from=builder /wifi-voucher-bot /wifi-voucher-bot

USER nonroot:nonroot
ENTRYPOINT ["/wifi-voucher-bot"]