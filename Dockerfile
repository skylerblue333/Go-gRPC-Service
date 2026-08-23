FROM golang:1.25.0-alpine AS builder
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/sky-rpc-core .

FROM gcr.io/distroless/static-debian12:nonroot
COPY --from=builder /out/sky-rpc-core /sky-rpc-core
EXPOSE 8080 9090
USER nonroot:nonroot
ENTRYPOINT ["/sky-rpc-core"]
