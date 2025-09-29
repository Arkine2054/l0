# Система обработки заказов

Микросервис для обработки заказов с использованием PostgreSQL, Kafka и веб-интерфейса.

### Компоненты системы:

- Controller Layer - HTTP API для веб-интерфейса
- Logic Layer - Бизнес-логика приложения
- Repository Layer - Работа с базой данных и кэшем
- Kafka Consumer - Обработка сообщений из Kafka
- Cache (LRU) - Кэширование часто запрашиваемых данных
- Database (PostgreSQL) - Основное хранилище данных

## Технологический стек

### Backend
- Go 1.24.1 - Основной язык программирования
- Gin - HTTP веб-фреймворк
- PostgreSQL 15 - Реляционная база данных
- Apache Kafka - Система обмена сообщениями
- LRU Cache - Кэширование в памяти

### Инфраструктура
- Docker & Docker Compose - Контейнеризация
- golang-migrate - Миграции базы данных
- Zookeeper - Координация Kafka

### Тестирование
- testify - Библиотека для тестирования
- gomock - Генерация моков
- httptest - HTTP тестирование

### Дополнительные библиотеки
- confluent-kafka-go - Kafka клиент
- lib/pq - PostgreSQL драйвер
- go-playground/validator - Валидация данных
- hashicorp/golang-lru - LRU кэш
- brianvoe/gofakeit - Генерация тестовых данных

## Структура проекта

```
l0/
├── cmd/                          # Точки входа приложения
├── internal/                     # Внутренние пакеты
│   ├── app/                      # Инициализация приложения
│   │   └── app.go
│   ├── controller/               # HTTP контроллеры
│   │   ├── controller.go
│   │   ├── controller_test.go
│   │   ├── logic_interface.go
│   │   └── logic_mock.go
│   ├── kafka/                    # Kafka интеграция
│   │   ├── consumer.go
│   │   ├── consumer_test.go
│   │   ├── producer.go
│   │   └── producer_test.go
│   ├── logic/                    # Бизнес-логика
│   │   ├── logic.go
│   │   └── logic_test.go
│   ├── models/                   # Модели данных
│   │   └── types.go
│   ├── repo/                     # Репозиторий
│   │   ├── repo.go
│   │   ├── repo_mock.go
│   │   └── repo_test.go
│   ├── shutdown/                 # Graceful shutdown
│   │   └── shutdown.go
│   └── util/                     # Утилиты
│       ├── kafka.go
│       ├── kafka_test.go
│       ├── unmarshal.go
│       └── unmarshal_test.go
├── migrations/                   # Миграции БД
│   ├── 1_init.up.sql
│   └── 1_init.down.sql
├── templates/                    # HTML шаблоны
│   ├── index.html
│   └── order.html
├── filler/                       # Генератор тестовых данных
│   ├── Dockerfile
│   └── main.go
├── vendor/                       # Зависимости
├── .env                         # Переменные окружения
├── .gitignore
├── docker-compose.yaml          # Docker Compose конфигурация
├── Dockerfile                   # Docker образ приложения
├── go.mod                       # Go модули
├── go.sum                       # Go зависимости
├── Makefile                     # Автоматизация задач
└── README.md                    # Документация
```

## Быстрый старт

### Предварительные требования

- Docker и Docker Compose
- Go 1.24.1+ (для локальной разработки)

### Запуск через Docker Compose

```bash
# Запуск всех сервисов
docker-compose up -d

# Просмотр логов
docker-compose logs -f app

# Остановка сервисов
docker-compose down
```


### Запуск тестов

```bash
# Все тесты
go test ./... -v

# Тесты с покрытием
go test ./... -cover

# Конкретный пакет
go test ./internal/logic/... -v

# Тесты с профилированием
go test ./... -cpuprofile=cpu.prof -memprofile=mem.prof
```

## Docker

### Docker Compose сервисы

- app - Основное приложение
- db - PostgreSQL база данных
- kafka - Apache Kafka
- zookeeper - Zookeeper для Kafka
- migrate - Миграции базы данных
- filler - Генератор тестовых данных
