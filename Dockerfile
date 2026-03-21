FROM golang:1.25@sha256:dfae680962532eeea67ab297f1166c2c4e686edb9a8f05f9d02d96fc9191833e AS builder

WORKDIR /
COPY go.mod go.sum ./
RUN go mod download

COPY *.go ./
COPY pkgs ./pkgs
RUN CGO_ENABLED=0 GOOS=linux go build -o /vault-init -trimpath .

FROM gcr.io/distroless/base:latest@sha256:347a41e7f263ea7f7aba1735e5e5b1439d9e41a9f09179229f8c13ea98ae94cf
WORKDIR /
COPY --from=builder /vault-init /vault-init
ENTRYPOINT ["/vault-init"]
