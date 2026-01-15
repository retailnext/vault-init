FROM golang:1.25@sha256:bc45dfd319e982dffe4de14428c77defe5b938e29d9bc6edfbc0b9a1fc171cb3 AS builder

WORKDIR /
COPY go.mod go.sum ./
RUN go mod download

COPY *.go ./
COPY pkgs ./pkgs
RUN CGO_ENABLED=0 GOOS=linux go build -o /vault-init -trimpath .

FROM gcr.io/distroless/base:latest@sha256:0c70ab46409b94a96f4e98e32e7333050581e75f7038de2877a4bfc146dfc7ce
WORKDIR /
COPY --from=builder /vault-init /vault-init
ENTRYPOINT ["/vault-init"]
