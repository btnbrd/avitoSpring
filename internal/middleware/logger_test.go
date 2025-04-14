package middleware

import (
	"bytes"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap/zapcore"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestInitLogger(t *testing.T) {
	logger, err := InitLogger()
	require.NoError(t, err)
	require.NotNil(t, logger)
}

func TestLoggingMiddleware(t *testing.T) {
	var buf bytes.Buffer
	writer := zapcore.AddSync(&buf)

	// Создаём кастомный логгер, чтобы перехватывать вывод
	encoderCfg := zap.NewProductionEncoderConfig()
	core := zapcore.NewCore(
		zapcore.NewJSONEncoder(encoderCfg),
		writer,
		zap.InfoLevel,
	)
	logger := zap.New(core)

	// Создаём gin с нашим middleware
	r := gin.New()
	r.Use(Logging(logger))
	r.GET("/ping", func(c *gin.Context) {
		c.String(200, "pong")
	})

	// Выполняем запрос
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/ping?foo=bar", nil)
	req.Header.Set("User-Agent", "unittest-agent")
	r.ServeHTTP(w, req)

	// Проверяем статус
	require.Equal(t, 200, w.Code)

	// Проверяем, что лог содержит нужные поля
	logOutput, err := io.ReadAll(&buf)
	require.NoError(t, err)
	require.Contains(t, string(logOutput), `"path":"/ping"`)
	require.Contains(t, string(logOutput), `"query":"foo=bar"`)
	require.Contains(t, string(logOutput), `"method":"GET"`)
	require.Contains(t, string(logOutput), `"user-agent":"unittest-agent"`)
}
