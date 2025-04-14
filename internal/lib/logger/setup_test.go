package logger_test

import (
	"log/slog"
	"testing"

	"github.com/stretchr/testify/assert"

	"pvz_service/internal/lib/logger"
)

func TestSetupLogger_LocalEnv(t *testing.T) {
	log := logger.SetupLogger("local")

	handler := log.Handler()
	assert.IsType(t, &slog.TextHandler{}, handler)
}

func TestSetupLogger_DevEnv(t *testing.T) {
	log := logger.SetupLogger("dev")

	handler := log.Handler()
	assert.IsType(t, &slog.JSONHandler{}, handler)
}

func TestSetupLogger_ProdEnv(t *testing.T) {
	log := logger.SetupLogger("prod")

	handler := log.Handler()
	assert.IsType(t, &slog.JSONHandler{}, handler)
}
