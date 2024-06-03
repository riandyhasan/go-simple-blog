# Build Stage
FROM golang:1.22-alpine AS builder

RUN apk update && apk add --no-cache git

WORKDIR /app

COPY . .
COPY ./migration.sql ./migration.sql
COPY ./pages/blogs.html ./blogs.html
COPY ./pages/blog-detail.html ./blog-detail.html
COPY ./pages/login.html ./login.html
COPY ./pages/register.html ./register.html

RUN go mod tidy

RUN go mod download

RUN CGO_ENABLED=0 GOOS=linux go build -o /app/go-simple-blog *.go

# Final Stage
FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /app

COPY --from=builder /app/go-simple-blog .
COPY --from=builder /app/migration.sql .
COPY --from=builder /app/blogs.html .
COPY --from=builder /app/blog-detail.html .
COPY --from=builder /app/login.html .
COPY --from=builder /app/register.html .

EXPOSE 8080

ENTRYPOINT ["./go-simple-blog"]