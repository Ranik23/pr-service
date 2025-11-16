FROM golang:1.25.1-alpine AS builder

RUN apk add --no-cache git ca-certificates

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

RUN go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main ./cmd/main

# Финальный образ
FROM alpine:latest

RUN apk add --no-cache ca-certificates

WORKDIR /app

COPY --from=builder /app/main .

COPY --from=builder /go/bin/migrate /usr/local/bin/migrate

COPY ./migrations ./migrations
COPY ./configs ./configs
COPY ./docs ./docs

EXPOSE 8080 9000

CMD ["./main"]

