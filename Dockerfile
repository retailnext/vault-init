FROM golang:1.22@sha256:c2010b9c2342431a24a2e64e33d9eb2e484af49e72c820e200d332d214d5e61f as builder

WORKDIR /
COPY go.mod go.sum ./
RUN go mod download

COPY *.go ./
COPY pkgs ./pkgs
RUN CGO_ENABLED=0 GOOS=linux go build -o /vault-init -trimpath .

FROM gcr.io/distroless/base:latest@sha256:786007f631d22e8a1a5084c5b177352d9dcac24b1e8c815187750f70b24a9fc6
WORKDIR /
COPY --from=builder /vault-init /vault-init
ENTRYPOINT ["/vault-init"]
