APP_NAME = l0

# Устанавливаем goimports
.PHONY: deps
deps:
	go install golang.org/x/tools/cmd/goimports@latest

# Форматирование кода
.PHONY: fmt
fmt:
	go fmt ./...
	goimports -w .

# Запуск линтера
.PHONY: lint
lint:
	golangci-lint run ./...

# Сборка бинаря
.PHONY: build
build:
	go build -o bin/$(APP_NAME) ./cmd/...

# Запуск сервиса локально
.PHONY: run
run:
	go run ./cmd/...

# Тесты
.PHONY: test
test:
	go test ./... -v
