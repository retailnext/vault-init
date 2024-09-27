FROM golang:1.23@sha256:8f6a7d881c5348114a19a08f8cc052b3a0d5341539e7ecc9e00902776949bc71 AS builder

WORKDIR /
COPY go.mod go.sum ./
RUN go mod download

COPY *.go ./
COPY pkgs ./pkgs
RUN CGO_ENABLED=0 GOOS=linux go build -o /vault-init -trimpath .

FROM gcr.io/distroless/base:latest@sha256:6ae5fe659f28c6afe9cc2903aebc78a5c6ad3aaa3d9d0369760ac6aaea2529c8
WORKDIR /
COPY --from=builder /vault-init /vault-init
ENTRYPOINT ["/vault-init"]
