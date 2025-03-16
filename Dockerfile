FROM golang:1.22.4

RUN go version

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY ./ ./

RUN rm -rf crypto-project
RUN go build -o crypto-project ./cmd/main.go
RUN chmod +x /app/crypto-project

CMD ["./crypto-project"]