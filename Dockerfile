FROM golang:1.25@sha256:b2663efc6db2bacaeda18f8a79f2e6e16c54ef694fb5dd726dbd3179741d6c52 AS builder

WORKDIR /
COPY go.mod go.sum ./
RUN go mod download

COPY *.go ./
COPY pkgs ./pkgs
RUN CGO_ENABLED=0 GOOS=linux go build -o /vault-init -trimpath .

FROM gcr.io/distroless/base:latest@sha256:9e9b50d2048db3741f86a48d939b4e4cc775f5889b3496439343301ff54cdba8
WORKDIR /
COPY --from=builder /vault-init /vault-init
ENTRYPOINT ["/vault-init"]
