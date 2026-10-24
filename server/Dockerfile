# syntax=docker/dockerfile:1
FROM golang:1.25-alpine AS builder
WORKDIR /app
ARG GOPROXY=https://proxy.golang.org,direct
COPY server/go.mod server/go.sum ./server/
RUN --mount=type=secret,id=ca_bundle \
  if [ -f /run/secrets/ca_bundle ]; then export SSL_CERT_FILE=/run/secrets/ca_bundle; fi; \
  cd server && GOPROXY="$GOPROXY" go mod download
COPY server/ ./server/
COPY docs/ ./docs/
RUN --mount=type=secret,id=ca_bundle \
  if [ -f /run/secrets/ca_bundle ]; then export SSL_CERT_FILE=/run/secrets/ca_bundle; fi; \
  cd server && CGO_ENABLED=0 go build -o /app/server ./cmd/main.go
RUN cd server && CGO_ENABLED=0 go build -o /app/click-worker ./cmd/click-worker
RUN cd server && CGO_ENABLED=0 go build -o /app/clickhouse-migrate ./cmd/clickhouse-migrate

FROM alpine:3.20
WORKDIR /app
COPY --from=builder /app/server .
COPY --from=builder /app/click-worker .
COPY --from=builder /app/clickhouse-migrate .
COPY --from=builder /app/docs ./docs
COPY --from=builder /app/server/migrations ./migrations
EXPOSE 8080
CMD ["./server"]
