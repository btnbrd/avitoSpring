package services

import (
	"avitoSpring/internal/config"
	"github.com/gin-gonic/gin"
)

type Server struct {
	*gin.Engine
	cfg *config.Config
}

func NewServer(cfg *config.Config) *Server {
	r := gin.New()
	r.Use(gin.Recovery())
	return &Server{Engine: r, cfg: cfg}
}

func (s *Server) GetConfig() *config.Config {
	return s.cfg
}
