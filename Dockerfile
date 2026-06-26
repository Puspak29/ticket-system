# build
FROM golang:1.26-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 go build -o /ticket-system .

# run
FROM alpine:3.23

WORKDIR /app
COPY --from=builder /ticket-system /app/ticket-system

EXPOSE 8080

ENTRYPOINT [ "/app/ticket-system" ]