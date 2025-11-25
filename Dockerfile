FROM golang:1.25-alpine AS builder
WORKDIR /app
COPY . .
RUN go build -o auth-service ./cmd

FROM alpine:latest
WORKDIR /app
COPY --from=builder /app/auth-service .
COPY --from=builder /app/config/local.yaml .
CMD ["./auth-service"]