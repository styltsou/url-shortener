FROM golang:1.25-alpine AS builder
WORKDIR /app
COPY server/go.mod server/go.sum ./server/
RUN cd server && go mod download
COPY server/ ./server/
COPY docs/ ./docs/
RUN CGO_ENABLED=0 go build -o /app/server ./server/cmd/main.go

FROM alpine:3.20
RUN apk add --no-cache ca-certificates
WORKDIR /app
COPY --from=builder /app/server .
COPY --from=builder /app/docs ./docs
COPY --from=builder /app/server/migrations ./migrations
EXPOSE 8080
CMD ["./server"]
