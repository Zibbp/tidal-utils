FROM golang:1.18 AS builder

RUN mkdir /app
ADD . /app
WORKDIR /app

RUN CGO_ENABLED=0 GOOS=linux go build -o app cmd/main.go
FROM alpine:latest AS production

RUN apk add --no-cache bash curl

COPY --from=builder /app .

COPY start.sh /exec/start.sh

RUN chmod +x /exec/start.sh && chmod +x app

CMD ["./exec/start.sh"]