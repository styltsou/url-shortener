FROM golang:1.25-alpine AS builder
WORKDIR /app
COPY server/go.mod server/go.sum ./server/
RUN cd server && go mod download
COPY server/ ./server/
COPY docs/ ./docs/
RUN cd server && CGO_ENABLED=0 go build -o /app/server ./cmd/main.go

FROM alpine:3.20
RUN apk add --no-cache ca-certificates wget
WORKDIR /app
COPY --from=builder /app/server .
COPY --from=builder /app/docs ./docs
COPY --from=builder /app/server/migrations ./migrations
EXPOSE 8080
HEALTHCHECK --interval=30s --timeout=3s --start-period=10s --retries=3 \
  CMD wget -qO- http://127.0.0.1:8080/api/v1/health/ready || exit 1
CMD ["./server"]
