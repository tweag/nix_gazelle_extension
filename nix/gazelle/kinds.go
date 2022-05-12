package gazelle

import (
	"github.com/bazelbuild/bazel-gazelle/rule"
)

// Kinds returns a map of maps rule names (kinds) and information on how to
// match and merge attributes that may be found in rules of those kinds. All
// kinds of rules generated for this language may be found here.
func (*nixLang) Kinds() map[string]rule.KindInfo {
	return map[string]rule.KindInfo{
		EXPORT_RULE: {
			MatchAny:   false,
			MatchAttrs: []string{"name"},
		},
		MANIFEST_RULE: {
			MatchAttrs: []string{"name", "nix_file_deps"},
			MergeableAttrs: map[string]bool{
				"nix_file_deps": true,
			},
		},
		PACKAGE_RULE: {
			MatchAttrs: []string{"name", "nix_file_deps"},
			MergeableAttrs: map[string]bool{
				"nix_file_deps": true,
			},
		},
	}
}

// Loads returns .bzl files and symbols they define. Every rule generated by
// GenerateRules, now or in the past, should be loadable from one of these
// files.
func (*nixLang) Loads() []rule.LoadInfo {
	return []rule.LoadInfo{
		{
			Name: "@io_tweag_gazelle_nix//nix:defs.bzl",
			Symbols: []string{
				EXPORT_RULE,
				MANIFEST_RULE,
			},
		},
		{
			Name: "@io_tweag_rules_nixpkgs//nixpkgs:nixpkgs.bzl",
			Symbols: []string{
				PACKAGE_RULE,
			},
		},
	}
}
