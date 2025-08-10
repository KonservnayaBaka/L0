FROM golang:1.24-alpine AS builder

RUN apk update && apk add --no-cache git

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN ls -la ./cmd .
RUN ls -la ./internal/db/migrations
RUN go build -o wbL0 ./cmd

FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /app

COPY --from=builder /app/wbL0 .
COPY --from=builder /app/internal/config/config.yml ./internal/config/config.yml
COPY --from=builder /app/internal/db/migrations ./internal/db/migrations

RUN chmod +x wbL0

CMD ["./wbL0"]