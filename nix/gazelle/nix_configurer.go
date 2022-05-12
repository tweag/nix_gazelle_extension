package gazelle

import (
	"flag"
	"strings"

	"github.com/bazelbuild/bazel-gazelle/config"
	"github.com/bazelbuild/bazel-gazelle/rule"
	"github.com/rs/zerolog"
	"github.com/tweag/nix_gazelle_extension/nix/gazelle/nixconfig"
	"github.com/tweag/nix_gazelle_extension/nix/gazelle/private/logconfig"
)

// Guarantee NixConfigurer implements Configurer interface
var (
	_ config.Configurer = &NixConfigurer{}
)

type NixConfigurer struct {
	logger *zerolog.Logger
}

func NewNixConfigurer() *NixConfigurer {
	return &NixConfigurer{
		logger: logconfig.GetLogger(),
	}
}

// RegisterFlags registers command-line flags used by the
// extension. This method is called once with the root configuration
// when Gazelle starts. RegisterFlags may set an initial values in
// Config.Exts. When flags are set, they should modify these values.
func (nlc NixConfigurer) RegisterFlags(
	flagSet *flag.FlagSet,
	cmd string,
	config *config.Config,
) {
}

func (nlc *NixConfigurer) CheckFlags(
	flagSet *flag.FlagSet,
	config *config.Config,
) error {
	return nil
}

// KnownDirectives returns a list of directive keys that this
// Configurer can interpret. Gazelle prints errors for directives that
// are not recoginized by any Configurer.
func (nlc *NixConfigurer) KnownDirectives() []string {
	return []string{
		nixconfig.NIX_PRELUDE,
		nixconfig.NIX_REPOSITORIES,
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
func (nlc *NixConfigurer) Configure(config *config.Config, relative string, buildFile *rule.File) {
	nlc.logger.
		Debug().
		Str("step", "gazelle.nixLang.Configurer.Configure").
		Str("path", relative).
		Msg("")

	// root config
	if _, exists := config.Exts[LANGUAGE_NAME]; !exists {
		config.Exts[LANGUAGE_NAME] = nixconfig.NixLanguageConfigs{
			"": nixconfig.New(),
		}
	}

	nixConfigs := config.Exts[LANGUAGE_NAME].(nixconfig.NixLanguageConfigs)
	_, exists := nixConfigs[relative]

	if !exists {
		nlc.logger.Trace().Msg("creating config")
		parent := nixConfigs.FindPackageParent(relative)
		nixConfigs[relative] = parent.NewChild()
	}
	if buildFile != nil {
		for _, directive := range buildFile.Directives {
			nlc.logger.Trace().
				Str("directive", directive.Key).
				Str("value", directive.Value).
				Msgf("setting config %s, using value %s", directive.Key, directive.Value)
			switch directive.Key {
			case nixconfig.NIX_PRELUDE:
				nixConfigs[relative].NixPrelude = directive.Value
			case nixconfig.NIX_REPOSITORIES:
				parseNixRepositories(nlc.logger, nixConfigs[relative], directive.Value)
			}
		}
	}
}

func parseNixRepositories(logger *zerolog.Logger, nixConfig *nixconfig.NixLanguageConfig, value string) {
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
			Str("directive", nixconfig.NIX_REPOSITORIES).
			Msgf("Cannot parse %s directive, invalid value %s", nixconfig.NIX_REPOSITORIES, value)
	}

	repositories := make(map[string]string)
	for i := 0; i < len(keyValuePair); i += 2 {
		repositories[keyValuePair[i]] = keyValuePair[i+1]
		nixConfig.NixRepositories = repositories
	}
}
