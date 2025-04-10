package handlers

import (
	"avitoSpring/internal/services"
	"github.com/gin-gonic/gin"
	"net/http"
)

// DummyLoginHandler обрабатывает запросы на /dummyLogin
func DummyLoginHandler(c *gin.Context) {
	c.String(http.StatusOK, "Этот эндпоинт принимает POST-запрос с параметром 'role' и возвращает токен.")
}

// RegisterHandlers регистрирует все обработчики
func RegisterHandlers(s *services.Server) {
	s.POST("/dummyLogin", DummyLoginHandler)
	s.POST("/register", RegisterHandler)
	s.POST("/login", LoginHandler)
	s.POST("/pvz", PvzHandler)
	s.GET("/pvz", PvzGetHandler)
	s.POST("/receptions", ReceptionHandler)
	s.POST("/products", ProductHandler)
}

// Пример регистрации других обработчиков

func RegisterHandler(c *gin.Context) {
	c.String(http.StatusOK, "Этот эндпоинт принимает данные для регистрации пользователя и возвращает созданного пользователя.")
}

func LoginHandler(c *gin.Context) {
	c.String(http.StatusOK, "Этот эндпоинт принимает данные для авторизации пользователя и возвращает токен.")
}

func PvzHandler(c *gin.Context) {
	c.String(http.StatusOK, "Этот эндпоинт создает новый ПВЗ. Доступ только для модераторов.")
}

func PvzGetHandler(c *gin.Context) {
	c.String(http.StatusOK, "Этот эндпоинт получает список ПВЗ с возможностью фильтрации по дате и пагинации.")
}

func ReceptionHandler(c *gin.Context) {
	c.String(http.StatusOK, "Этот эндпоинт создает новую приемку товаров для ПВЗ.")
}

func ProductHandler(c *gin.Context) {
	c.String(http.StatusOK, "Этот эндпоинт добавляет товар в текущую приемку для ПВЗ.")
}
