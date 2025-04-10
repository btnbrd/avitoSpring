package services

import (
	"github.com/gin-gonic/gin"
)

type Server struct {
	*gin.Engine
}

// NewServer создает новый сервер с логированием
func NewServer() *Server {
	r := gin.Default()
	return &Server{r}
}
