# Используем официальный образ Go 1.23
FROM golang:1.23

# Устанавливаем рабочую директорию внутри контейнера
WORKDIR /app

# Копируем все файлы в рабочую директорию контейнера
COPY . .

# Скачиваем зависимости и собираем приложение
RUN go mod tidy && go build -o /build ./cmd/app/main.go && go clean -cache -modcache

# Открываем порт 8080 для связи
EXPOSE 8080

# Указываем команду для запуска
CMD ["/build"]


# Используем минимальный образ для запуска
#FROM gcr.io/distroless/base
#
#WORKDIR /root/
#
## Копируем собранный исполнимый файл
#COPY --from=build /app/main .
#
## Запускаем Go-сервис
#CMD ["./main"]

