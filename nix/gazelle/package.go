package gazelle

import (
	"github.com/bazelbuild/bazel-gazelle/rule"
)

// FIXME: nixPackage should not carry information
// meant for build context target(s) i.e. nix_export
type NixRuleArgs struct {
	attrs    map[string]interface{}
	kind     string
	comments []string
}

func genNixRule(args *NixRuleArgs) *rule.Rule {
	r := rule.NewRule(args.kind, "")
	for k, v := range args.attrs {
		r.SetAttr(k, v)
	}
	for _, c := range args.comments {
		r.AddComment(c)
	}
	return r
}
