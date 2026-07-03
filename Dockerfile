FROM golang:1.22-alpine AS builder

WORKDIR /app

COPY go.mod ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /lime ./cmd/lime

FROM alpine:3.19

WORKDIR /app
COPY --from=builder /lime /app/lime

EXPOSE 8080
ENTRYPOINT ["/app/lime"]
