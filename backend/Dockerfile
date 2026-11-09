# Build stage
FROM golang:1.25-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o bin/server ./cmd/server \
    && CGO_ENABLED=0 GOOS=linux go build -o bin/migrator ./cmd/migrator

# Runtime stage
FROM alpine:3.21

RUN adduser -D -u 10001 appuser

WORKDIR /app

COPY --from=builder /app/bin/server /app/bin/migrator /app/bin/
COPY --from=builder /app/config-yaml /app/config-yaml
COPY --from=builder /app/migrations /app/migrations

RUN chown -R appuser:appuser /app
USER appuser

EXPOSE 8080

# CONFIG_PATH selects the yaml config (config-yaml/prod.yaml by default)
ENV CONFIG_PATH=/app/config-yaml/prod.yaml

CMD ["/app/bin/server"]
