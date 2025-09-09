FROM golang:1.25@sha256:b773c94fd926723a415b28e615016a49793764e270566a82c6a47afb4c52f7de AS builder

WORKDIR /
COPY go.mod go.sum ./
RUN go mod download

COPY *.go ./
COPY pkgs ./pkgs
RUN CGO_ENABLED=0 GOOS=linux go build -o /vault-init -trimpath .

FROM gcr.io/distroless/base:latest@sha256:fa15492938650e1a5b87e34d47dc7d99a2b4e8aefd81b931b3f3eb6bb4c1d2f6
WORKDIR /
COPY --from=builder /vault-init /vault-init
ENTRYPOINT ["/vault-init"]
