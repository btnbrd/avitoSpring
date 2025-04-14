package handlers

import "avitoSpring/internal/services"

func RegisterHandlers(s *services.Server, authService *services.AuthService, pvzService *services.PVZService, receptionService *services.ReceptionService, productService *services.ProductService) {

	authHandler := NewAuthHandler(authService)
	pvzHandler := NewPVZHandler(pvzService, authHandler)
	receptionHandler := NewReceptionHandler(receptionService, authHandler)
	productHandler := NewProductHandler(productService, authHandler)

	s.POST("/dummyLogin", authHandler.DummyLoginHandler)
	s.POST("/register", authHandler.RegisterHandler)
	s.POST("/login", authHandler.LoginHandler)
	s.POST("/pvz", authHandler.AuthMiddleware(), pvzHandler.PvzHandler)
	s.GET("/pvz", authHandler.AuthMiddleware(), pvzHandler.PvzGetHandler)
	s.POST("/receptions", authHandler.AuthMiddleware(), receptionHandler.ReceptionHandler)
	s.POST("/products", authHandler.AuthMiddleware(), productHandler.ProductHandler)

	s.POST("/pvz/:pvzId/close_last_reception", authHandler.AuthMiddleware(), receptionHandler.CloseLastReceptionHandler)
	s.POST("/pvz/:pvzId/delete_last_product", authHandler.AuthMiddleware(), productHandler.DeleteLastProductHandler)
}
