package server

import (
	"avitoSpring/internal/config"
	"github.com/gin-gonic/gin"
	"reflect"
	"testing"
)

func TestNewServer(t *testing.T) {
	cfg := &config.Config{
		Server: config.ServerConfig{
			Port: ":8080",
		},
		Database: config.DatabaseConfig{
			Host:     "localhost",
			Port:     5432,
			Name:     "testdb",
			User:     "testuser",
			Password: "testpass",
		},
	}

	server := NewServer(cfg)

	if server == nil {
		t.Fatal("Expected non-nil server, got nil")
	}

	if server.Engine == nil {
		t.Fatal("Expected non-nil gin.Engine, got nil")
	}

	if reflect.TypeOf(server.Engine) != reflect.TypeOf(&gin.Engine{}) {
		t.Errorf("Expected Engine to be of type *gin.Engine, got %T", server.Engine)
	}

	if server.cfg != cfg {
		t.Errorf("Expected cfg to be %v, got %v", cfg, server.cfg)
	}

	registeredMiddlewares := server.Engine.Handlers
	if len(registeredMiddlewares) == 0 {
		t.Error("Expected at least one middleware (Recovery), got none")
	}
	server.GET("/test-panic", func(c *gin.Context) {
		panic("test panic")
	})

}

func TestServer_GetConfig(t *testing.T) {
	// Создаем тестовую конфигурацию
	cfg := &config.Config{
		Server: config.ServerConfig{
			Port: ":8080",
		},
		Database: config.DatabaseConfig{
			Host:     "localhost",
			Port:     5432,
			Name:     "testdb",
			User:     "testuser",
			Password: "testpass",
		},
	}

	server := NewServer(cfg)

	returnedCfg := server.GetConfig()

	if returnedCfg != cfg {
		t.Errorf("Expected config %v, got %v", cfg, returnedCfg)
	}

	if returnedCfg.Server.Port != cfg.Server.Port {
		t.Errorf("Expected server port %q, got %q", cfg.Server.Port, returnedCfg.Server.Port)
	}
	if returnedCfg.Database.Host != cfg.Database.Host ||
		returnedCfg.Database.Port != cfg.Database.Port ||
		returnedCfg.Database.Name != cfg.Database.Name ||
		returnedCfg.Database.User != cfg.Database.User ||
		returnedCfg.Database.Password != cfg.Database.Password {
		t.Errorf("Expected database config %v, got %v", cfg.Database, returnedCfg.Database)
	}
}
