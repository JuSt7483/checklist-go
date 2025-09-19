# Этап сборки
FROM golang:1.24-alpine AS builder

WORKDIR /app

# Копируем файлы модулей и загружаем зависимости
COPY go.mod go.sum ./
RUN go mod download

# Копируем весь исходный код проекта в контекст сборки
COPY . .

# Собираем приложение api-service
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/api-service ./services/api-service/main.go

# Этап выполнения
FROM alpine:latest

WORKDIR /app

# Копируем бинарный файл из этапа сборки
COPY --from=builder /app/api-service .

# Открываем порт HTTP
EXPOSE 8080

# Команда для запуска сервиса
CMD ["/app/api-service"]