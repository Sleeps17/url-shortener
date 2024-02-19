FROM golang:1.22

WORKDIR /app

COPY go.mod ./
COPY go.sum ./

RUN go mod download

COPY . .

RUN go build -o url-shortener ./cmd/url-shortener

ENV CONFIG_PATH=./config/local.yaml

EXPOSE 8080

CMD ["./url-shortener"]
