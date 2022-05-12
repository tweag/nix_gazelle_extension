package gazelle

import (
	"flag"
	"sort"
	"strings"

	"github.com/bazelbuild/bazel-gazelle/config"
	"github.com/bazelbuild/bazel-gazelle/language"
	golang "github.com/bazelbuild/bazel-gazelle/language/go"
	"github.com/bazelbuild/bazel-gazelle/language/proto"
	"github.com/bazelbuild/bazel-gazelle/rule"
	"github.com/bazelbuild/bazel-gazelle/walk"
	"github.com/rs/zerolog"
	"github.com/tweag/nix_gazelle_extension/nix/gazelle/nixconfig"
)

func (nixLang *nixLang) UpdateRepos(
	args language.UpdateReposArgs,
) language.UpdateReposResult {
	logger := nixLang.logger.With().
		Str("step", "gazelle.nixLang.UpdateRepos").
		Str("path", args.Config.WorkDir).
		Str("language", LANGUAGE_NAME).
		Logger()

	logger.Debug().Msg("")

	rules := collectDependenciesFromRepo(&logger, args.Config, nixLang)

	sortRules(rules)
	return language.UpdateReposResult{
		Error: nil,
		Gen:   rules,
	}
}

func sortRules(rules []*rule.Rule) {
	sort.SliceStable(rules, func(i, j int) bool {
		if cmp := strings.Compare(rules[i].Name(), rules[j].Name()); cmp != 0 {
			return cmp < 0
		}
		return rules[i].AttrString("name") < rules[j].AttrString("name")
	})
}

func collectDependenciesFromRepo(
	logger *zerolog.Logger,
	extensionConfig *config.Config,
	lang language.Language,
) []*rule.Rule {
	rules := make([]*rule.Rule, 0)

	cexts := []config.Configurer{
		&config.CommonConfigurer{},
		&walk.Configurer{},
		lang,
		golang.NewLanguage(),
		proto.NewLanguage(),
	}

	initUpdateReposConfig(logger, extensionConfig, cexts)

	walk.Walk(
		extensionConfig,
		cexts,
		[]string{},
		walk.VisitAllUpdateDirsMode,
		func(
			_,
			_ string,
			_ *config.Config,
			_ bool,
			buildFile *rule.File,
			_,
			_,
			_ []string,
		) {
			// Translate to repository rules.
			if buildFile != nil {
				for _, ruleStatement := range buildFile.Rules {
					if ruleStatement.Kind() == MANIFEST_RULE {
						// Change rule kind to include required load statements
						// in WORKSPACE file
						ruleStatement.SetKind(PACKAGE_RULE)
						rules = append(rules, ruleStatement)
					}
				}
			}
		},
	)

	return rules
}

func initUpdateReposConfig(logger *zerolog.Logger, extensionConfig *config.Config, cexts []config.Configurer) {
	// root config
	if _, exists := extensionConfig.Exts[LANGUAGE_NAME]; !exists {
		extensionConfig.Exts[LANGUAGE_NAME] = nixconfig.NixLanguageConfigs{
			"": nixconfig.New(),
		}
	}

	extensionConfig.Exts[LANGUAGE_NAME].(nixconfig.NixLanguageConfigs)[""].Wsmode = true

	flagSet := flag.NewFlagSet("updateReposFlagSet", flag.ContinueOnError)

	for _, cext := range cexts {
		cext.RegisterFlags(flagSet, "update", extensionConfig)
	}

	for _, cext := range cexts {
		if err := cext.CheckFlags(flagSet, extensionConfig); err != nil {
			logger.Fatal().
				Err(err).
				Msg("")
		}
	}
}
