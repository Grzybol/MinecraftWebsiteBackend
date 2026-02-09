FROM golang:1.22-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o backend ./

FROM alpine:3.19

RUN addgroup -S app && adduser -S app -G app

WORKDIR /app

COPY --from=builder /app/backend ./backend
COPY docker-entrypoint.sh /usr/local/bin/docker-entrypoint.sh
RUN chmod +x /usr/local/bin/docker-entrypoint.sh

USER app

EXPOSE 8080

ENTRYPOINT ["docker-entrypoint.sh"]
