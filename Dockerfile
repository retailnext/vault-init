FROM golang:1.23@sha256:95983c2237525489841f6c9dc8e0df3d5cfbcf7788d74a6720ef5163ab7fc0f7 AS builder

WORKDIR /
COPY go.mod go.sum ./
RUN go mod download

COPY *.go ./
COPY pkgs ./pkgs
RUN CGO_ENABLED=0 GOOS=linux go build -o /vault-init -trimpath .

FROM gcr.io/distroless/base:latest@sha256:74ddbf52d93fafbdd21b399271b0b4aac1babf8fa98cab59e5692e01169a1348
WORKDIR /
COPY --from=builder /vault-init /vault-init
ENTRYPOINT ["/vault-init"]
