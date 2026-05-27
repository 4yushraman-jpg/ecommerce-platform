FROM golang:1.25-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /bin/auth-service ./cmd/main.go

FROM alpine:3.22
RUN adduser -D -g '' appuser
WORKDIR /app
COPY --from=builder /bin/auth-service /usr/local/bin/auth-service
EXPOSE 8080
USER appuser
HEALTHCHECK --interval=30s --timeout=5s --start-period=15s --retries=3 CMD wget -qO- http://localhost:8080/health || exit 1
ENTRYPOINT ["/usr/local/bin/auth-service"]
