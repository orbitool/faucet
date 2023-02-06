# Build stage
FROM golang:1.19.2-alpine AS builder
RUN apk add --no-cache gcc
RUN apk add --no-cache musl-dev

WORKDIR /usr/src/app

COPY go.mod ./
COPY go.sum ./
RUN go mod download && go mod verify

COPY . .
RUN go build -v -o /usr/local/bin/faucet ./cli/main.go

# Run stage
FROM alpine:latest
RUN apk --no-cache add ca-certificates

WORKDIR /usr/src/app

COPY --from=builder /usr/local/bin/faucet /usr/local/bin/faucet
CMD ["faucet", "--captcha"]