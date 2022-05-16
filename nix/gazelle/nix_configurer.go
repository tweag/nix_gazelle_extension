package gazelle

import (
	"errors"
	"flag"
	"strings"

	"github.com/bazelbuild/bazel-gazelle/config"
	"github.com/bazelbuild/bazel-gazelle/rule"
	"github.com/lainio/err2"
	"github.com/lainio/err2/try"
	"github.com/rs/zerolog"
	"github.com/tweag/nix_gazelle_extension/nix/gazelle/nixconfig"
	"github.com/tweag/nix_gazelle_extension/nix/gazelle/private/logconfig"
)

// Guarantee NixConfigurer implements Configurer interface
var (
	_         config.Configurer = &NixConfigurer{}
	errAssert                   = errors.New("assertion failed")
	errParse                    = errors.New("directive parsing failed")
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

	nlc.logger.Trace().Msg("creating config")

	cfg := createNixConfig(config, relative)
	var directive rule.Directive
	var dk, dv string

	defer err2.Catch(func(err error) {
		nlc.logger.Fatal().
			Err(err).
			Str("value", dk).
			Str("directive", dv).
			Msgf("Cannot parse %s directive, invalid value %s", dk, dv)
	})

	if buildFile != nil {
		for _, directive = range buildFile.Directives {
			dk, dv = directive.Key, directive.Value
			nlc.logger.Trace().
				Str("directive", dk).
				Str("value", dv).
				Msgf("setting config %s, using value %s", dk, dv)
			switch directive.Key {
			case nixconfig.NIX_PRELUDE:
				try.To(parseNixPrelude(cfg, dv))
			case nixconfig.NIX_REPOSITORIES:
				try.To(parseNixRepositories(cfg, dv))
			}
		}
	}
}

func parseNixPrelude(nixConfig *nixconfig.NixLanguageConfig, value string) (err error) {
	//TODO implmement parsing
	nixConfig.NixPrelude = value
	return nil
}

func parseNixRepositories(nixConfig *nixconfig.NixLanguageConfig, value string) (err error) {
	const parts = 2

	pairs := strings.Split(value, " ")
	keyValuePair := make([]string, 0, len(pairs)*parts)

	for _, key := range pairs {
		keyValuePair = append(keyValuePair, strings.Split(key, "=")...)
	}

	if len(keyValuePair)%2 != 0 {
		return errParse
	}

	repositories := make(map[string]string)
	for i := 0; i < len(keyValuePair); i += 2 {
		repositories[keyValuePair[i]] = keyValuePair[i+1]
		nixConfig.NixRepositories = repositories
	}

	return nil
}

func GetNixConfig(config *config.Config, relative string) (*nixconfig.NixLanguageConfig, error) {
	configs, ok := config.Exts[LANGUAGE_NAME].(nixconfig.NixLanguageConfigs)
	if !ok {
		return nil, errAssert
	}

	cfg, ok := configs[relative]
	if !ok {
		return nil, errAssert
	}

	return cfg, nil
}

func createNixConfig(config *config.Config, relative string) *nixconfig.NixLanguageConfig {
	var ok bool
	var cfg *nixconfig.NixLanguageConfig
	var cfgs nixconfig.NixLanguageConfigs

	if cfgs, ok = config.Exts[LANGUAGE_NAME].(nixconfig.NixLanguageConfigs); !ok {
		cfgs = nixconfig.NixLanguageConfigs{
			"": nixconfig.New(),
		}
		config.Exts[LANGUAGE_NAME] = cfgs
	}

	if cfg, ok = cfgs[relative]; !ok {
		cfg = cfgs.FindPackageParent(relative).NewChild()
		cfgs[relative] = cfg
	}

	return cfg
}
