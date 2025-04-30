FROM golang:1.23-alpine

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go get github.com/rabbitmq/amqp091-go
RUN go get github.com/valkey-io/valkey-go  # Cambiado de go-redis a valkey-go

RUN go build -o consumer .

CMD ["./consumer"]