FROM golang:1.21-alpine AS builder
WORKDIR /app
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o grpc-service main.go
FROM alpine:latest
WORKDIR /app
COPY --from=builder /app/grpc-service .
EXPOSE 8080
CMD ["./grpc-service"]
