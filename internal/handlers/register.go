package handlers

import (
	"avitoSpring/internal/middleware"
	"avitoSpring/internal/server"
	"avitoSpring/internal/services"
)

func RegisterHandlers(s *server.Server, authService *services.AuthService,
	pvzService *services.PVZService, receptionService *services.ReceptionService,
	productService *services.ProductService, jwtService *services.JWTService) {

	authHandler := NewAuthHandler(authService)
	pvzHandler := NewPVZHandler(pvzService, authHandler)
	receptionHandler := NewReceptionHandler(receptionService, authHandler)
	productHandler := NewProductHandler(productService, authHandler)
	authorizer := middleware.NewAuthorizer(jwtService)
	s.POST("/dummyLogin", authHandler.DummyLoginHandler)
	s.POST("/register", authHandler.RegisterHandler)
	s.POST("/login", authHandler.LoginHandler)
	s.POST("/pvz", authorizer.AuthMiddleware(), pvzHandler.PvzHandler)
	s.GET("/pvz", authorizer.AuthMiddleware(), pvzHandler.PvzGetHandler)
	s.POST("/receptions", authorizer.AuthMiddleware(), receptionHandler.ReceptionHandler)
	s.POST("/products", authorizer.AuthMiddleware(), productHandler.ProductHandler)

	s.POST("/pvz/:pvzId/close_last_reception", authorizer.AuthMiddleware(), receptionHandler.CloseLastReceptionHandler)
	s.POST("/pvz/:pvzId/delete_last_product", authorizer.AuthMiddleware(), productHandler.DeleteLastProductHandler)
}
