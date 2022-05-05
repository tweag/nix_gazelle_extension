package gazelle

import (
	"flag"
	"strings"

	"github.com/bazelbuild/bazel-gazelle/config"
	"github.com/bazelbuild/bazel-gazelle/rule"
	"github.com/rs/zerolog"
	"github.com/tweag/nix_gazelle_extension/nix/gazelle/nixconfig"
)

type Configurer struct {
	lang *nixLang
}

func NewConfigurer(lang *nixLang) *Configurer {
	return &Configurer{lang}
}

// RegisterFlags registers command-line flags used by the
// extension. This method is called once with the root configuration
// when Gazelle starts. RegisterFlags may set an initial values in
// Config.Exts. When flags are set, they should modify these values.
func (nixLangConfigurer Configurer) RegisterFlags(
	flagSet *flag.FlagSet,
	cmd string,
	config *config.Config,
) {
}

func (nixLangConfigurer *Configurer) CheckFlags(
	flagSet *flag.FlagSet,
	config *config.Config,
) error {
	return nil
}

// KnownDirectives returns a list of directive keys that this
// Configurer can interpret. Gazelle prints errors for directives that
// are not recoginized by any Configurer.
func (nixLangConfigurer *Configurer) KnownDirectives() []string {
	return []string{
		nixconfig.NixPrelude,
		nixconfig.NixRepositories,
	}
}

// Configure modifies the configuration using directives and other
// information extracted from a build file. Configure is called in
// each directory.
//
// c is the configuration for the current directory. It starts out as
// a copy of the configuration for the parent directory.
//
// rel is the slash-separated relative path from the repository root
// to the current directory. It is "" for the root directory itself.
//
// f is the build file for the current directory or nil if there is no
// existing build file.

func (nixLangConfigurer *Configurer) Configure(config *config.Config, relative string, buildFile *rule.File) {
	logger := nixLangConfigurer.lang.logger.With().
		Str("step", "gazelle.nixLang.Configurer.Configure").
		Str("path", relative).
		Logger()

	logger.Debug().Msg("")

	// root config
	if _, exists := config.Exts[languageName]; !exists {
		config.Exts[languageName] = nixconfig.Configs{
			"": nixconfig.New(),
		}
	}

	nixConfigs := config.Exts[languageName].(nixconfig.Configs)
	cfg, exists := nixConfigs[relative]

	if !exists {
		logger.Trace().Msg("creating config")
		parent := nixConfigs.ParentForPackage(relative)
		cfg = parent.NewChild()
		nixConfigs[relative] = cfg
	}
	if buildFile != nil {
		for _, directive := range buildFile.Directives {
			logger.Trace().
				Str("directive", directive.Key).
				Str("value", directive.Value).
				Msgf("setting config %s, using value %s", directive.Key, directive.Value)
			switch directive.Key {
			case nixconfig.NixPrelude:
				nixConfigs[relative].SetNixPrelude(directive.Value)
			case nixconfig.NixRepositories:
				parseNixRepositories(&logger, nixConfigs[relative], directive.Value)
			}
		}
	}
}

func parseNixRepositories(logger *zerolog.Logger, nixConfig *nixconfig.Config, value string) {
	const parts = 2

	pairs := strings.Split(value, " ")
	keyValuePair := make([]string, 0, len(pairs)*parts)

	for _, key := range pairs {
		keyValuePair = append(keyValuePair, strings.Split(key, "=")...)
	}

	if len(keyValuePair)%2 != 0 {
		logger.Panic().
			Err(errParse).
			Str("value", value).
			Str("directive", nixconfig.NixRepositories).
			Msgf("Cannot parse %s directive, invalid value %s", nixconfig.NixRepositories, value)
	}

	repositories := make(map[string]string)
	for i := 0; i < len(keyValuePair); i += 2 {
		repositories[keyValuePair[i]] = keyValuePair[i+1]
		nixConfig.SetNixRepositories(repositories)
	}
}
