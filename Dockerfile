FROM golang:1.24-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o /tgbot ./cmd/tgbot

FROM alpine:3.18

RUN apk --no-cache add ca-certificates

COPY --from=builder /tgbot /tgbot

RUN chmod +x /tgbot

ENTRYPOINT ["/tgbot"]