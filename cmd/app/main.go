package main

import (
	"avitoSpring/internal/handlers"
	"avitoSpring/internal/middleware"
	"avitoSpring/internal/services"
	//"fmt"
	"log"
)

func main() {
	logger := middleware.InitLogger()
	defer logger.Sync()

	// Создаем сервер
	s := services.NewServer()

	// Регистрируем middleware для логирования
	s.Use(middleware.Logging(logger))

	// Регистрируем обработчики
	handlers.RegisterHandlers(s)

	// Запускаем серверды
	log.Fatal(s.Run(":8080"))
}
