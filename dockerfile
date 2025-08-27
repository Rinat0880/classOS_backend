# Простый Dockerfile без PowerShell
# PowerShell команды будут выполняться на Windows Server хосте

FROM golang:1.24-alpine AS builder

WORKDIR /app

# Копируем зависимости и код
COPY go.mod go.sum ./
RUN go mod download

COPY . .

# Собираем приложение
RUN CGO_ENABLED=0 GOOS=linux go build -o main cmd/main.go

# Финальный минимальный образ
FROM alpine:3.19

# Устанавливаем только необходимые утилиты
RUN apk add --no-cache ca-certificates tzdata curl

# Создаем пользователя
RUN adduser -D -s /bin/sh appuser

WORKDIR /app

# Копируем приложение
COPY --from=builder /app/main .
COPY --from=builder /app/configs ./configs

RUN chown -R appuser:appuser /app

USER appuser

EXPOSE 8000

CMD ["./main"]