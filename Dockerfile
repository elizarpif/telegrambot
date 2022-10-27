# syntax=docker/dockerfile:1
FROM --platform=linux/amd64 golang:1.19-alpine

WORKDIR /app

COPY . ./
RUN go mod download

RUN go build github.com/elizarpif/telegrambot/cmd/nonsense

CMD ./nonsense