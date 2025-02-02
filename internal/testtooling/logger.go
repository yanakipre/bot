package testtooling

import "github.com/yanakipe/bot/internal/logger"

// SetNewGlobalLoggerQuietly sets global logger once per application and does not panic if already
// set.
// Not flexible at all. Useful for tests, when you need to somehow set everything up.
func SetNewGlobalLoggerQuietly() {
	cfg := logger.DefaultConfig()
	cfg.Format = logger.FormatConsole
	cfg.LogLevel = "DEBUG"
	logger.SetNewGlobalLoggerQuietly(cfg)
}
