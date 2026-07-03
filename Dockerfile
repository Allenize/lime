# --- Build stage ---
FROM golang:1.22-alpine AS builder

WORKDIR /app

# Cache dependencies first
COPY go.mod ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /lime ./cmd/lime

# --- Run stage (small final image) ---
FROM alpine:3.19

WORKDIR /app
COPY --from=builder /lime /app/lime

EXPOSE 8080
ENTRYPOINT ["/app/lime"]
