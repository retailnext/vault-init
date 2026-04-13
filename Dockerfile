FROM golang:1.26@sha256:5e69504bb541a36b859ee1c3daeb670b544b74b3dfd054aea934814e9107f6eb AS builder

WORKDIR /
COPY go.mod go.sum ./
RUN go mod download

COPY *.go ./
COPY pkgs ./pkgs
RUN CGO_ENABLED=0 GOOS=linux go build -o /vault-init -trimpath .

FROM gcr.io/distroless/base:latest@sha256:c83f022002fc917a92501a8c30c605efdad3010157ba2c8998a2cbf213299201
WORKDIR /
COPY --from=builder /vault-init /vault-init
ENTRYPOINT ["/vault-init"]
