FROM golang:latest

WORKDIR /app
COPY . .

RUN go mod tidy

CMD ["make", "run"]
