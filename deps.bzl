load("@io_bazel_rules_go//go:deps.bzl", "go_rules_dependencies")
load("//third_party/go:repositories.bzl", "go_deps")
load(
    "@io_tweag_rules_nixpkgs//nixpkgs:repositories.bzl",
    "rules_nixpkgs_dependencies",
)

def io_tweag_gazelle_nix_deps():
    """ Load dependencies required by dependencies of gazelle nix. """
    go_rules_dependencies()
    rules_nixpkgs_dependencies()
    go_deps()
