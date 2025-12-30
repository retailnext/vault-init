FROM golang:1.25@sha256:b6ba5234ed128185b0b81070813c77dd5c973aec7703f4646805f11377627408 AS builder

WORKDIR /
COPY go.mod go.sum ./
RUN go mod download

COPY *.go ./
COPY pkgs ./pkgs
RUN CGO_ENABLED=0 GOOS=linux go build -o /vault-init -trimpath .

FROM gcr.io/distroless/base:latest@sha256:f5a3067027c2b322cd71b844f3d84ad3deada45ceb8a30f301260a602455070e
WORKDIR /
COPY --from=builder /vault-init /vault-init
ENTRYPOINT ["/vault-init"]
