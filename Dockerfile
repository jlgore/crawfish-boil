FROM golang:1.24-alpine AS builder
RUN apk add --no-cache git
WORKDIR /build
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o honeypot ./cmd/honeypot

FROM scratch
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /build/honeypot /honeypot
USER 65534:65534
EXPOSE 18789
ENTRYPOINT ["/honeypot"]
