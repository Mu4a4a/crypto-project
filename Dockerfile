FROM golang:1.23.7-alpine3.21 AS builder

WORKDIR /app

COPY . ./

RUN go mod download

RUN go build -o /app/cryptoproject ./cmd/main.go

CMD ["/app/cryptoproject"]

FROM alpine:3.21 as runner

WORKDIR /app

COPY --from=builder /app/cryptoproject /app/
COPY --from=builder /app/migrations /app/migrations
COPY ./config/config.yaml ./.env /app/

RUN chmod +x /app/cryptoproject

CMD ["/app/cryptoproject"]
