package logconfig

import (
	"fmt"
	"os"
	"strings"

	"github.com/rs/zerolog"
)

const logLevelKey = "GAZELLE_LANGUAGES_NIX_LOG_LEVEL"

func LogLevel() (zerolog.Level) {
	v := os.Getenv(logLevelKey)
	if v == "" {
		v = "info"
	}

	logLevel, err := zerolog.ParseLevel(strings.ToLower(v))
	if err != nil {
		panic(fmt.Sprintf("invalid log level '%s': %s", v, err))
	}

	return logLevel
}
