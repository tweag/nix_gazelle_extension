package gazelle

import (
	"github.com/bazelbuild/bazel-gazelle/rule"
)

var nixKinds = map[string]rule.KindInfo{
	exportRule: {
		MatchAny:   false,
		MatchAttrs: []string{"name"},
	},
	manifestRule: {
		MatchAttrs: []string{"name", "nix_file_deps"},
		MergeableAttrs: map[string]bool{
			"nix_file_deps": true,
		},
	},
	packageRule: {
		MatchAttrs: []string{"name", "nix_file_deps"},
		MergeableAttrs: map[string]bool{
			"nix_file_deps": true,
		},
	},
}

var nixLoads = []rule.LoadInfo{
	{
		Name: "@io_tweag_gazelle_nix//nix:defs.bzl",
		Symbols: []string{
			exportRule,
			manifestRule,
		},
	},
	{
		Name: "@io_tweag_rules_nixpkgs//nixpkgs:nixpkgs.bzl",
		Symbols: []string{
			packageRule,
		},
	},
}

// Kinds returns a map of maps rule names (kinds) and information on how to
// match and merge attributes that may be found in rules of those kinds. All
// kinds of rules generated for this language may be found here.
func (_ *nixLang) Kinds() map[string]rule.KindInfo { return nixKinds }

// Loads returns .bzl files and symbols they define. Every rule generated by
// GenerateRules, now or in the past, should be loadable from one of these
// files.
func (_ *nixLang) Loads() []rule.LoadInfo { return nixLoads }
