FROM golang:1.23@sha256:8956c08c8129598db36e92680d6afda0079b6b32b93c2c08260bf6fa75524e07 AS builder

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
