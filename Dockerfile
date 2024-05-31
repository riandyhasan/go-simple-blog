# Build Stage
FROM golang:1.22-alpine AS builder

RUN apk update && apk add --no-cache git

WORKDIR /app

COPY . .
COPY ./migration.sql ./migration.sql

RUN go mod tidy

RUN go mod download

RUN CGO_ENABLED=0 GOOS=linux go build -o /app/go-simple-blog *.go

# Final Stage
FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /app

COPY --from=builder /app/go-simple-blog .
COPY --from=builder /app/migration.sql .

EXPOSE 8080

ENTRYPOINT ["./go-simple-blog"]