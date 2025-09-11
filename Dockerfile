# syntax=docker/dockerfile:1.18
FROM golang:1.25-alpine AS builder

ARG VERSION=dev
ARG COMMIT=dirty
ARG LDFLAGS="-s -w -X main.Version=${VERSION} -X main.Commit=${COMMIT}"
ENV CGO_ENABLED=0

WORKDIR /workspace
COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go build -ldflags="${LDFLAGS}" -o /out/vex ./cmd/vex/main.go

# Prepare world-writable dirs for arbitrary UID (OpenShift)
RUN mkdir -p /outfs/work /outfs/tmp && chmod 0777 /outfs/work /outfs/tmp

# Final: FROM scratch
FROM scratch AS final
COPY --from=builder /out/vex /vex
COPY --from=builder /outfs/work /work
COPY --from=builder /outfs/tmp  /tmp
ENV HOME=/tmp
WORKDIR /work
ENTRYPOINT ["/vex"]

