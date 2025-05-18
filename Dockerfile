FROM golang:1.23.8-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go build -ldflags="-s -w" -o app ./example

FROM alpine:latest

COPY --from=builder /app/app /app/app


WORKDIR /app

EXPOSE 8080
CMD ["./app"]
