# ── Stage 1: build ────────────────────────────────────────────
FROM golang:1.26-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o /bot ./cmd/bot/

# ── Stage 2: runtime ──────────────────────────────────────────
FROM alpine:3.21

WORKDIR /app

COPY --from=builder /bot .

ENTRYPOINT ["./bot", "--config", "/app/configs/config.yaml"]
