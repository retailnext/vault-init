FROM golang:1.22@sha256:1b70fa8200ed48367e7d0e0e82c0c1d14cf712d422116f9d69e7c964243653a0 as builder

WORKDIR /
COPY go.mod go.sum ./
RUN go mod download

COPY *.go ./
COPY pkgs ./pkgs
RUN CGO_ENABLED=0 GOOS=linux go build -o /vault-init -trimpath .

FROM gcr.io/distroless/base:latest@sha256:1aae189e3baecbb4044c648d356ddb75025b2ba8d14cdc9c2a19ba784c90bfb9
WORKDIR /
COPY --from=builder /vault-init /vault-init
ENTRYPOINT ["/vault-init"]
