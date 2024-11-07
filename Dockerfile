FROM golang:1.23@sha256:d56c3e08fe5b27729ee3834854ae8f7015af48fd651cd25d1e3bcf3c19830174 AS builder

WORKDIR /
COPY go.mod go.sum ./
RUN go mod download

COPY *.go ./
COPY pkgs ./pkgs
RUN CGO_ENABLED=0 GOOS=linux go build -o /vault-init -trimpath .

FROM gcr.io/distroless/base:latest@sha256:8fe31fb9d159141d9c3ff99f1fd287239d89d97ea95fea1f08f82ea5f2b544da
WORKDIR /
COPY --from=builder /vault-init /vault-init
ENTRYPOINT ["/vault-init"]
