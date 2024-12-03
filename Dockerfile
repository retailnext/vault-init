FROM golang:1.23@sha256:b4aabba13d38c069fa6d6ee3656e1fcb11c18181ecfd13728bfced414f6422ca AS builder

WORKDIR /
COPY go.mod go.sum ./
RUN go mod download

COPY *.go ./
COPY pkgs ./pkgs
RUN CGO_ENABLED=0 GOOS=linux go build -o /vault-init -trimpath .

FROM gcr.io/distroless/base:latest@sha256:e9d0321de8927f69ce20e39bfc061343cce395996dfc1f0db6540e5145bc63a5
WORKDIR /
COPY --from=builder /vault-init /vault-init
ENTRYPOINT ["/vault-init"]
