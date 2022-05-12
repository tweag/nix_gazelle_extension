package logconfig

import (
	"os"
	"strings"
	"sync"

	"github.com/rs/zerolog"
)

const GAZELLE_NIX_LOGGING_LEVEL = "GAZELLE_LANGUAGES_NIX_LOG_LEVEL"
const DEFAULT_LOGGING_LEVEL = "info"

var loggerInstance *zerolog.Logger
var once sync.Once

func getLoggingLevel() zerolog.Level {
	var verbosity string
	if verbosity = os.Getenv(GAZELLE_NIX_LOGGING_LEVEL); len(verbosity) <= 0 {
		verbosity = DEFAULT_LOGGING_LEVEL
	}

	logLevel, err := zerolog.ParseLevel(strings.ToLower(verbosity))
	if err != nil {
		// Gracefully handle invalid user settings
		return zerolog.InfoLevel
	}

	return logLevel
}

// Return instance of zerolog logger
// that should be used throughout Gazelle nix execution
func GetLogger() *zerolog.Logger {
	once.Do(func() {
		var logger = zerolog.New(
			zerolog.ConsoleWriter{Out: os.Stderr},
		).With().
			Timestamp().
			Caller().
			Logger().
			Level(getLoggingLevel())
		loggerInstance = &logger
	})
	return loggerInstance
}
