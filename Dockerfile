FROM golang:1.24@sha256:ceb568d0de81fbef40ce4fee77eab524a0a0a8536065c51866ad8c59b7a912cf AS builder

WORKDIR /
COPY go.mod go.sum ./
RUN go mod download

COPY *.go ./
COPY pkgs ./pkgs
RUN CGO_ENABLED=0 GOOS=linux go build -o /vault-init -trimpath .

FROM gcr.io/distroless/base:latest@sha256:125eb09bbd8e818da4f9eac0dfc373892ca75bec4630aa642d315ecf35c1afb7
WORKDIR /
COPY --from=builder /vault-init /vault-init
ENTRYPOINT ["/vault-init"]
