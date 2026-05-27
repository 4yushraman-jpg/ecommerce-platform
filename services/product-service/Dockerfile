FROM golang:1.25-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /bin/product-service ./cmd/main.go

FROM alpine:3.22
RUN adduser -D -g '' appuser
WORKDIR /app
COPY --from=builder /bin/product-service /usr/local/bin/product-service
EXPOSE 8081
USER appuser
HEALTHCHECK --interval=30s --timeout=5s --start-period=15s --retries=3 CMD wget -qO- http://localhost:8081/health || exit 1
ENTRYPOINT ["/usr/local/bin/product-service"]
