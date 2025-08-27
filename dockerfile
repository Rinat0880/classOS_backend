FROM golang:1.24-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o main cmd/main.go

FROM alpine:3.19

RUN apk add --no-cache ca-certificates tzdata curl

RUN adduser -D -s /bin/sh appuser

WORKDIR /app

COPY --from=builder /app/main .
COPY --from=builder /app/configs ./configs
COPY --from=builder /app/.env .env* ./

RUN chown -R appuser:appuser /app

USER appuser

EXPOSE 8000

CMD ["./main"]