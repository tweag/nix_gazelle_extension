package gazelle

import (
	"github.com/bazelbuild/bazel-gazelle/config"
	"github.com/bazelbuild/bazel-gazelle/label"
	"github.com/bazelbuild/bazel-gazelle/repo"
	"github.com/bazelbuild/bazel-gazelle/resolve"
	"github.com/bazelbuild/bazel-gazelle/rule"
)

type Resolver struct {
	lang *nixLang
}

func NewResolver(lang *nixLang) *Resolver {
	return &Resolver{
		lang: lang,
	}
}

func (Resolver) Name() string {
	return languageName
}

// Imports returns a list of ImportSpecs that can be used to import
// the rule r. This is used to populate RuleIndex.
//
// If nil is returned, the rule will not be indexed. If any non-nil
// slice is returned, including an empty slice, the rule will be
// indexed.
func (nixLangResolver Resolver) Imports(
	extensionConfig *config.Config,
	ruleStatement *rule.Rule,
	buildFile *rule.File,
) []resolve.ImportSpec {
	logger := nixLangResolver.lang.logger.With().
		Str("step", "gazelle.nixLang.Resolver.Imports").
		Str("path", buildFile.Pkg).
		Str("rule", ruleStatement.Name()).
		Logger()

	logger.Debug().Msg("")

	var prefix string

	switch ruleStatement.Kind() {
	case exportRule:
		prefix = "exports:"
	case packageRule:
		prefix = "nixpkgs_package:"
	}

	return []resolve.ImportSpec{{
		Lang: languageName,
		Imp:  prefix + ruleStatement.Name(),
	}}
}

// Embeds returns a list of labels of rules that the given rule
// embeds. If a rule is embedded by another importable rule of the
// same language, only the embedding rule will be indexed. The
// embedding rule will inherit the imports of the embedded rule.
func (Resolver) Embeds(r *rule.Rule, from label.Label) []label.Label {
	return nil
}

// Resolve translates imported libraries for a given rule into Bazel
// dependencies. A list of imported libraries is typically stored in a
// private attribute of the rule when it's generated (this interface
// doesn't dictate how that is stored or represented). Resolve
// generates a "deps" attribute (or the appropriate language-specific
// equivalent) for each import according to language-specific rules
// and heuristics.
func (Resolver) Resolve(
	config *config.Config,
	ruleIndex *resolve.RuleIndex,
	remoteCache *repo.RemoteCache,
	rule *rule.Rule,
	importsRaw interface{},
	from label.Label,
) {
}
