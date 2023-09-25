FROM golang:1.20

WORKDIR /app

ENV MYAPP=prod

COPY go.mod go.sum ./

RUN go mod download

COPY . .

COPY config/ /app/config/

RUN go build -o main .

CMD ["./main"]
