# --- Этап 1: Сборщик (Builder) ---
# Используем официальный образ Go на базе Alpine.
# Alpine - легковесный дистрибутив, что делает итоговый образ меньше.
FROM golang:1.24-alpine AS builder

# Устанавливаем рабочую директорию внутри контейнера
WORKDIR /app

# Устанавливаем утилиту для миграций 'migrate'.
# Она нужна для накатывания изменений схемы базы данных.
RUN go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest

# Копируем файлы управления зависимостями
COPY go.mod go.sum ./
# Загружаем зависимости. Этот шаг кэшируется Docker,
# если файлы go.mod/go.sum не менялись, что ускоряет сборку.
RUN go mod download

# Копируем все исходные коды проекта в контейнер
COPY . .

# Собираем наше Go-приложение в один бинарный файл.
# CGO_ENABLED=0 отключает Cgo, GOOS=linux собирает под Linux.
# Результат - статический бинарник /app/db-service.
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/db-service ./services/db-service/main.go

# --- Этап 2: Финальный образ ---
# Используем самый базовый образ Alpine. Он очень маленький (~5MB).
FROM alpine:latest

WORKDIR /app

# Копируем утилиту migrate из сборщика.
COPY --from=builder /go/bin/migrate /usr/local/bin/

# Копируем скомпилированное приложение из сборщика.
COPY --from=builder /app/db-service .

# Копируем файлы миграций, чтобы утилита migrate могла их найти.
COPY services/db-service/migrations ./migrations

# Открываем порт, который слушает наш gRPC-сервис.
EXPOSE 50051

# Команда по умолчанию для запуска сервиса.
# В docker-compose.yml мы ее переопределим, чтобы сначала запустить миграции.
CMD ["/app/db-service"]