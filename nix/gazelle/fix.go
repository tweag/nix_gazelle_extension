package gazelle

import (
	"strings"

	"github.com/bazelbuild/bazel-gazelle/config"
	"github.com/bazelbuild/bazel-gazelle/rule"
)

// Fix repairs deprecated usage of language-specific rules in f. This is
// called before the file is indexed. Unless c.ShouldFix is true, fixes
// that delete or rename rules should not be performed.
func (nixLang *nixLang) Fix(c *config.Config, buildFile *rule.File) {
	for _, loadStatement := range buildFile.Loads {
		if loadStatement.Has(MANIFEST_RULE) {
			loadStatement.Remove(MANIFEST_RULE)

			if loadStatement.IsEmpty() {
				loadStatement.Delete()
			}
		}
	}

	var knownRuleStatements []*rule.Rule

	for _, ruleStatement := range buildFile.Rules {
		if ruleStatement.Kind() == MANIFEST_RULE {
			knownRuleStatements = append(knownRuleStatements, ruleStatement)
			continue
		}
		if ruleStatement.Kind() == EXPORT_RULE && strings.HasSuffix(ruleStatement.AttrString("name"), "-exports") {
			knownRuleStatements = append(knownRuleStatements, ruleStatement)
		}
	}

	for _, knownRuleStatement := range knownRuleStatements {
		knownRuleStatement.Delete()
	}
}
