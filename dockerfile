FROM golang:1.24-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o main cmd/main.go

RUN CGO_ENABLED=0 GOOS=linux go build -o test-ldap cmd/ldap-test/main.go

FROM alpine:3.19

RUN apk add --no-cache ca-certificates tzdata curl openssl

RUN adduser -D -s /bin/sh appuser

WORKDIR /app

COPY --from=builder /app/main .
COPY --from=builder /app/test-ldap .
COPY --from=builder /app/configs ./configs
COPY --from=builder /app/.env .env* ./

COPY certs/ad-ca.cer /usr/local/share/ca-certificates/ad-ca.cer

RUN update-ca-certificates

RUN chown -R appuser:appuser /app

USER appuser

EXPOSE 8000

CMD ["./main"]