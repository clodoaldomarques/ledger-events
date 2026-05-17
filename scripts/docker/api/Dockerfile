## build
FROM golang:1.26.1 AS build

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
COPY cmd/main.go .
RUN go build -o /ledger-events

## deploy
FROM ubuntu AS base

WORKDIR /

COPY --from=build /ledger-events /ledger-events

EXPOSE 5001

ENTRYPOINT [ "/ledger-events" ]