FROM golang:1.22-alpine AS builder

LABEL MAINTAINER="sleeps17"

WORKDIR /go/src/app

RUN apk add upx

RUN apk --no-cache add git bash make gcc musl-dev

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go build -ldflags="-s -w" -o /go/bin/url-shortener ./cmd/url-shortener
RUN upx -9 /go/bin/url-shortener

FROM alpine:latest AS runner

COPY --from=builder /go/bin/url-shortener ./
COPY config/dev.yaml config/dev.yaml

ENV CONFIG_PATH=/config/dev.yaml

EXPOSE 8081

CMD ["./url-shortener"]
