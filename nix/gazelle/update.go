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
)

func (nixLang *nixLang) UpdateRepos(
	args language.UpdateReposArgs,
) language.UpdateReposResult {
	logger := nixLang.logger.With().
		Str("step", "gazelle.nixLang.UpdateRepos").
		Str("path", args.Config.WorkDir).
		Str("language", languageName).
		Logger()

	logger.Debug().Msg("")

	packageList := collectDependenciesFromRepo(&logger, args.Config, nixLang)
	sortRules(packageList)
	return language.UpdateReposResult{
		Error: nil,
		Gen:   packageList,
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
	cexts := []config.Configurer{
		&config.CommonConfigurer{},
		&walk.Configurer{},
		lang,
		golang.NewLanguage(),
		proto.NewLanguage(),
	}

	initUpdateReposConfig(logger, extensionConfig, cexts)

	var result []*rule.Rule
	walk.Walk(extensionConfig, cexts, []string{}, walk.VisitAllUpdateDirsMode, func(dir, rel string, c *config.Config, update bool, f *rule.File, subdirs, regularFiles, genFiles []string) {
		// Generate rules.
		var empty, gen []*rule.Rule
		var imports []interface{}
		res := lang.GenerateRules(language.GenerateArgs{
			Config:       c,
			Dir:          dir,
			Rel:          rel,
			File:         f,
			Subdirs:      subdirs,
			RegularFiles: regularFiles,
			GenFiles:     genFiles,
			OtherEmpty:   empty,
			OtherGen:     gen})
		if len(res.Gen) != len(res.Imports) {
			logger.Panic().Msgf("%s: language %s generated %d rules but returned %d imports", rel, languageName, len(res.Gen), len(res.Imports))
		}
		empty = append(empty, res.Empty...)
		gen = append(gen, res.Gen...)
		imports = append(imports, res.Imports...)
		if f == nil && len(gen) == 0 {
			return
		}

		for _, ruleStatement := range res.Gen {
			if ruleStatement.Kind() == packageRule {
				result = append(result, ruleStatement)
			}
		}
	},
	)

	return result
}

func initUpdateReposConfig(logger *zerolog.Logger, extensionConfig *config.Config, cexts []config.Configurer) {
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
