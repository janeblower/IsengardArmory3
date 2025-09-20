# Этап сборки
FROM golang:1.25-alpine AS builder
WORKDIR /app
RUN apk add --no-cache git
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -o isengard-armory .   # <- изменили здесь

# Финальный образ
FROM alpine:latest
RUN apk add --no-cache ca-certificates
WORKDIR /root/
COPY --from=builder /app/isengard-armory .
CMD ["./isengard-armory"]
