FROM golang:1.23@sha256:b2ca38170893394183f940a7f988bf15c4112a4ddb73214941fe4d08a09f9329 AS builder

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
