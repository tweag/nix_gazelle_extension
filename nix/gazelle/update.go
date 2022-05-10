package gazelle

import (
	"flag"
	"path/filepath"
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
	kinds := make(map[string]rule.KindInfo)

	initUpdateReposConfig(logger, extensionConfig, cexts)

	var result []*rule.Rule

	walk.Walk(extensionConfig, cexts, []string{}, walk.VisitAllUpdateDirsMode, func(dir, rel string, c *config.Config, _ bool, f *rule.File, subdirs, regularFiles, genFiles []string) {
		// Generate rules.
		var empty, gen []*rule.Rule
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
		gen = append(gen, res.Gen...)
		if f == nil && len(gen) == 0 {
			return
		}

		// Apply and record relevant kind mappings.
		var (
			mappedKindInfo = make(map[string]rule.KindInfo)
		)
		for _, r := range gen {
			if repl, ok := c.KindMap[r.Kind()]; ok {
				mappedKindInfo[repl.KindName] = kinds[r.Kind()]
				r.SetKind(repl.KindName)
			}
		}

		// Insert or merge rules into the build file.
		if f == nil {
			f = rule.EmptyFile(filepath.Join(dir, c.DefaultBuildFileName()), rel)
		}

		// TODO: support merges
		for _, ruleStatement := range res.Gen {
			if ruleStatement.Kind() == packageRule {
				result = append(result, ruleStatement)
			} else {
				ruleStatement.Insert(f)
			}
		}

		f.Save(f.Path)
	},
	)

	return result
}

func initUpdateReposConfig(logger *zerolog.Logger, extensionConfig *config.Config, cexts []config.Configurer) {
	// root config
	if _, exists := extensionConfig.Exts[languageName]; !exists {
		extensionConfig.Exts[languageName] = nixconfig.Configs{
			"": nixconfig.New(),
		}
	}

	extensionConfig.Exts[languageName].(nixconfig.Configs)[""].SetWsMode(true)

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
