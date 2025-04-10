# Используем официальный образ Go 1.23
FROM golang:1.23 AS build

# Устанавливаем рабочую директорию в контейнере
WORKDIR /app

# Копируем все файлы проекта в контейнер
COPY . .

# Загружаем зависимости
RUN go mod tidy

# Собираем приложение
RUN go build -o main .

# Используем минимальный образ для запуска
FROM gcr.io/distroless/base

WORKDIR /root/

# Копируем собранный исполнимый файл
COPY --from=build /app/main .

# Запускаем Go-сервис
CMD ["./main"]

