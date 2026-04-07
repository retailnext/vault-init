FROM golang:1.26@sha256:5e69504bb541a36b859ee1c3daeb670b544b74b3dfd054aea934814e9107f6eb AS builder

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
