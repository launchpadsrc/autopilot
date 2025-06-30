FROM golang:1.24.2 AS builder

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY . .

RUN go build -o main .

FROM gcr.io/distroless/base

WORKDIR /app

COPY --from=builder /app/main .

COPY --from=builder /app/bot.yml .

COPY --from=builder /app/bot/locales ./bot/locales

COPY --from=builder /app/ai.yml .

COPY --from=builder /app/database/migrations ./database/migrations

CMD ["./main"]
