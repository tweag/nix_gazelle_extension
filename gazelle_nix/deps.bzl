load("@io_bazel_rules_go//go:deps.bzl", "go_rules_dependencies")
load(
    "@io_tweag_rules_nixpkgs//nixpkgs:repositories.bzl",
    "rules_nixpkgs_dependencies",
)

def io_tweag_gazelle_nix_deps():
    """ Load packages required by gazelle_nix extension. """
    go_rules_dependencies()
    rules_nixpkgs_dependencies()
