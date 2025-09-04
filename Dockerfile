FROM golang:1.25@sha256:5e856b892fff06368bd92fe3ac00c457ede6d399e7152575b57f9a14312ab713 AS builder

WORKDIR /
COPY go.mod go.sum ./
RUN go mod download

COPY *.go ./
COPY pkgs ./pkgs
RUN CGO_ENABLED=0 GOOS=linux go build -o /vault-init -trimpath .

FROM gcr.io/distroless/base:latest@sha256:d605e138bb398428779e5ab490a6bbeeabfd2551bd919578b1044718e5c30798
WORKDIR /
COPY --from=builder /vault-init /vault-init
ENTRYPOINT ["/vault-init"]
