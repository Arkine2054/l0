# build stage
FROM golang:1.24.1 AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -o service main.go

# get migrate binary
FROM golang:1.24.1 AS migrate-builder
RUN go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest

# final image (bookworm чтобы не было проблем с glibc)
FROM debian:bookworm-slim
WORKDIR /app

COPY --from=builder /app/service .
COPY --from=migrate-builder /go/bin/migrate /usr/local/bin/migrate
COPY migrations ./migrations
COPY templates ./templates

CMD migrate -path ./migrations -database "postgres://$DB_USER:$DB_PASSWORD@$DB_HOST:$DB_PORT/$DB_NAME?sslmode=disable" up && ./service