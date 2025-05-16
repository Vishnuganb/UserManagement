# Build stage
FROM golang:1.24-alpine AS builder

WORKDIR /app

COPY . .

RUN go mod tidy
RUN go build -o main ./cmd/server/main.go

# Run stage
FROM alpine:3.19
WORKDIR /app
# Install postgresql-client
RUN apk add --no-cache postgresql-client

COPY --from=builder /app/main .
COPY app.env .
COPY start.sh .
COPY internal/db/migration ./internal/db/migration

EXPOSE 8080 8082
CMD [ "/app/main" ]
ENTRYPOINT [ "/app/start.sh" ]