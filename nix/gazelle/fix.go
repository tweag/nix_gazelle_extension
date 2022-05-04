package gazelle

import (
	"github.com/bazelbuild/bazel-gazelle/config"
	"github.com/bazelbuild/bazel-gazelle/rule"
)

// Fix repairs deprecated usage of language-specific rules in f. This is
// called before the file is indexed. Unless c.ShouldFix is true, fixes
// that delete or rename rules should not be performed.
func (nixLang *nixLang) Fix(c *config.Config, buildFile *rule.File) {
	for _, loadStatement := range buildFile.Loads {
		if loadStatement.Has(exportRule) {
			loadStatement.Remove(exportRule)

			if loadStatement.IsEmpty() {
				loadStatement.Delete()
			}
		}
	}

	var knownRuleStatements []*rule.Rule

	for _, ruleStatement := range buildFile.Rules {
		if ruleStatement.Kind() == exportRule {
			knownRuleStatements = append(knownRuleStatements, ruleStatement)
		}
	}

	for _, knownRuleStatement := range knownRuleStatements {
		knownRuleStatement.Delete()
	}
}
