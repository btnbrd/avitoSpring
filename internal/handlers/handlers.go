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
	authenticator := middleware.NewAuthenticator(jwtService)
	s.POST("/dummyLogin", authHandler.DummyLoginHandler)
	s.POST("/register", authHandler.RegisterHandler)
	s.POST("/login", authHandler.LoginHandler)
	s.POST("/pvz", authenticator.AuthMiddleware(), pvzHandler.PvzHandler)
	s.GET("/pvz", authenticator.AuthMiddleware(), pvzHandler.PvzGetHandler)
	s.POST("/receptions", authenticator.AuthMiddleware(), receptionHandler.ReceptionHandler)
	s.POST("/products", authenticator.AuthMiddleware(), productHandler.ProductHandler)

	s.POST("/pvz/:pvzId/close_last_reception", authenticator.AuthMiddleware(), receptionHandler.CloseLastReceptionHandler)
	s.POST("/pvz/:pvzId/delete_last_product", authenticator.AuthMiddleware(), productHandler.DeleteLastProductHandler)
}
