FROM golang:1.24-bookworm AS builder

WORKDIR /build
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o packllama ./cmd/packllama

FROM debian:bookworm-slim
RUN apt-get update && apt-get install -y --no-install-recommends ca-certificates wget && rm -rf /var/lib/apt/lists/*

WORKDIR /app
COPY --from=builder /build/packllama .

EXPOSE 8080
ENTRYPOINT ["/app/packllama"]
CMD ["--host", "0.0.0.0", "--port", "8080"]
