FROM golang:1.23@sha256:73f06be4578c9987ce560087e2e2ea6485fb605e3910542cadd8fa09fc5f3e31 AS builder

WORKDIR /
COPY go.mod go.sum ./
RUN go mod download

COPY *.go ./
COPY pkgs ./pkgs
RUN CGO_ENABLED=0 GOOS=linux go build -o /vault-init -trimpath .

FROM gcr.io/distroless/base:latest@sha256:7a4bffcb07307d97aa731b50cb6ab22a68a8314426f4e4428335939b5b1943a5
WORKDIR /
COPY --from=builder /vault-init /vault-init
ENTRYPOINT ["/vault-init"]
