package gazelle

import (
	"errors"

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
	logger *zerolog.Logger
}

// Return implementation supporting nix language
func NewLanguage() language.Language {
	logger := logconfig.GetLogger()
	logger.Debug().Msg("creating nix language")

	return &nixLang{
		logger:     logconfig.GetLogger(),
		Configurer: NewNixConfigurer(),
		Resolver:   NewNixResolver(),
	}
}
