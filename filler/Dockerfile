# build stage
FROM golang:1.24.1 AS builder
WORKDIR /app

# копируем go.mod и go.sum из корня
COPY go.mod go.sum ./
RUN go mod download

# копируем весь проект (чтобы видеть internal/)
COPY . .

# билдим только filler
WORKDIR /app/filler
RUN go build -o filler main.go

# final image
FROM debian:bookworm-slim
WORKDIR /app
COPY --from=builder /app/filler/filler .

CMD ["./filler"]
