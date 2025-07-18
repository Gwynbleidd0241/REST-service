FROM golang:1.23-alpine AS build-stage
LABEL authors="gwynbleidd"

WORKDIR /src

COPY go.mod go.sum ./
RUN go mod tidy

COPY . .
RUN go build -o app ./cmd/main.go

FROM alpine:latest

WORKDIR /app

COPY --from=build-stage /src/app .
COPY .env .

EXPOSE 8081
ENTRYPOINT ["./app"]
