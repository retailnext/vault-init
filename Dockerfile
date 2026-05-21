FROM golang:1.26@sha256:6df14f4a4bc9d979a3721f488981e0d1b318006377e473ed23d026796f5f4c0a AS builder

WORKDIR /
COPY go.mod go.sum ./
RUN go mod download

COPY *.go ./
COPY pkgs ./pkgs
RUN CGO_ENABLED=0 GOOS=linux go build -o /vault-init -trimpath .

FROM gcr.io/distroless/base:latest@sha256:f2df8702d4dcc45ce76df6cbc14ad1975fcf88a04bd0e8947b6194264f9ab75e
WORKDIR /
COPY --from=builder /vault-init /vault-init
ENTRYPOINT ["/vault-init"]
