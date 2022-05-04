package gazelle

import (
	"errors"
	"os"

	"github.com/bazelbuild/bazel-gazelle/config"
	"github.com/bazelbuild/bazel-gazelle/language"
	"github.com/bazelbuild/bazel-gazelle/resolve"
	"github.com/rs/zerolog"
	"github.com/tweag/nix_gazelle_extension/nix/gazelle/private/logconfig"
)

var (
	errAssert = errors.New("assertion failed")
	errParse  = errors.New("directive parsing failed")
)

type nixLang struct {
	config.Configurer
	resolve.Resolver

	logger zerolog.Logger
}

// NewLanguage implementation.
func NewLanguage() language.Language {

	logLevel := logconfig.LogLevel()
	logger := zerolog.New(zerolog.ConsoleWriter{Out: os.Stderr}).
		With().
		Timestamp().
		Caller().
		Logger().
		Level(logLevel)
	logger.Debug().Msg("creating nix language")

	nl := nixLang{
		logger: logger,
	}
	nl.Configurer = NewConfigurer(&nl)
	nl.Resolver = NewResolver(&nl)

	return &nl
}
