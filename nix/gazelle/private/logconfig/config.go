package logconfig

import (
	"fmt"
	"os"
	"regexp"
	"strings"
	"sync"

	"github.com/lainio/err2"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/pkgerrors"
)

const (
	COLOR_BLUE    = 34
	COLOR_MAGENTA = 35

	COLOR_BOLD = 1

	GAZELLE_NIX_LOGGING_LEVEL = "GAZELLE_LANGUAGES_NIX_LOG_LEVEL"
	DEFAULT_LOGGING_LEVEL     = "info"
)

var loggerInstance *zerolog.Logger
var once sync.Once

func getLoggingLevel() zerolog.Level {
	var verbosity string
	zerolog.ErrorStackMarshaler = pkgerrors.MarshalStack

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
		loglevel := getLoggingLevel()
		var logger = zerolog.New(
			zerolog.ConsoleWriter{Out: os.Stderr},
		).With().
			Timestamp().
			Caller().
			Logger().
			Level(loglevel)
		loggerInstance = &logger

		if loglevel == zerolog.TraceLevel {
			callerRegex, _ := regexp.Compile(`[a-z][^\\\\\:\|\<\>\"\*\?]*:\d+`)

			details := []*string{}
			cw := zerolog.ConsoleWriter{
				Out:              os.Stderr,
				FormatCaller:     callerFormatter(),
				FormatFieldName:  fieldNameFormatter(),
				FormatFieldValue: fieldValueFormatter(),
				FormatLevel:      levelFormatter(),
				PartsExclude:     []string{"message"},
			}

			err2.StackTraceWriter = loggerInstance.
				Output(cw).
				Hook(stackTraceHook(callerRegex, details))
		}
	})

	return loggerInstance
}

func callerFormatter() zerolog.Formatter {
	return func(i interface{}) string {
		return fmt.Sprintf("\x1b[%dm%-70s\x1b[0m|", COLOR_BOLD, i.(string))
	}
}

func fieldValueFormatter() zerolog.Formatter {
	return func(i interface{}) string {
		return fmt.Sprintf("\x1b[%dm%s\x1b[0m", COLOR_MAGENTA, i.(string))
	}
}

func fieldNameFormatter() zerolog.Formatter {
	return func(i interface{}) string {
		return ""
	}
}

func levelFormatter() zerolog.Formatter {
	return func(i interface{}) string {
		if i == nil {
			return fmt.Sprintf(
				"\x1b[%dm%v\x1b[0m",
				COLOR_MAGENTA, "STK",
			)
		}
		return i.(string)
	}
}

func stackTraceHook(re *regexp.Regexp, details []*string) zerolog.HookFunc {
	return func(e *zerolog.Event, _ zerolog.Level, msg string) {
		if strings.Contains(msg, "\n") {
			e.Discard()
			return
		}

		res := re.FindString(msg)
		if res != "" {
			e.Str("caller", res)
			e.Str("call", *details[len(details)-1])
			details = nil
			return
		}

		details = append(details, &msg)
		e.Discard()
	}
}
