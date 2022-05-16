package gazelle

import (
	"github.com/bazelbuild/bazel-gazelle/config"
	"github.com/bazelbuild/bazel-gazelle/language"
	"github.com/bazelbuild/bazel-gazelle/resolve"
	"github.com/rs/zerolog"
	"github.com/tweag/nix_gazelle_extension/nix/gazelle/private/logconfig"
)

// Guarantee nixLang implements Language, RepoUpdater interfaces
var (
	_ language.Language    = &nixLang{}
	_ language.RepoUpdater = &nixLang{}
)

type nixLang struct {
	config.Configurer
	resolve.Resolver
	logger *zerolog.Logger
}

// Return implementation supporting nix language
func NewLanguage() language.Language {
	logconfig.GetLogger().Debug().Msg("creating nix language")

	return &nixLang{
		logger:     logconfig.GetLogger(),
		Configurer: NewNixConfigurer(),
		Resolver:   NewNixResolver(),
	}
}
