package gazelle

import (
	"flag"
	"strings"

	"github.com/bazelbuild/bazel-gazelle/config"
	"github.com/bazelbuild/bazel-gazelle/rule"
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
func (nc Configurer) RegisterFlags(
	fs *flag.FlagSet,
	cmd string,
	c *config.Config,
) {
}

func (jc *Configurer) CheckFlags(
	fs *flag.FlagSet,
	c *config.Config,
) error {
	return nil
}

// KnownDirectives returns a list of directive keys that this
// Configurer can interpret. Gazelle prints errors for directives that
// are not recoginized by any Configurer.
func (nc *Configurer) KnownDirectives() []string {
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

func (nc *Configurer) Configure(c *config.Config, relative string, buildFile *rule.File) {
	log := nc.lang.logger.With().
		Str("step", "gazelle.nixLang.Configurer.Configure").
		Str("path", relative).
		Logger()

	log.Debug().Msg("")

	if _, exists := c.Exts[languageName]; !exists {
		c.Exts[languageName] = nixconfig.Configs{
			"": nixconfig.New(),
		}
	}
	cfgs := c.Exts[languageName].(nixconfig.Configs)
	if _, exists := cfgs[relative]; !exists {
		cfgs[relative] = nixconfig.New()
	}
	if buildFile != nil {
		for _, directive := range buildFile.Directives {
			switch directive.Key {
			case nixconfig.NixPrelude:
				cfgs[relative].SetNixPrelude(directive.Value)
			case nixconfig.NixRepositories:
				parseNixRepositories(nc, cfgs[relative], directive.Value)
			}
		}
	}
}

func parseNixRepositories(nc *Configurer, nixConfig *nixconfig.Config, value string) {
	const parts = 2

	pairs := strings.Split(value, " ")
	keyValuePair := make([]string, 0, len(pairs)*parts)

	for _, key := range pairs {
		keyValuePair = append(keyValuePair, strings.Split(key, "=")...)
	}

	if len(keyValuePair)%2 != 0 {
		nc.lang.logger.Panic().
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
