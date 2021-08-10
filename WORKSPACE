workspace(
    name = "io_tweag_gazelle_nix",
)

load("//:defs.bzl", "io_tweag_gazelle_nix_repositories")

io_tweag_gazelle_nix_repositories()

load(
    "@io_tweag_rules_nixpkgs//nixpkgs:repositories.bzl",
    "rules_nixpkgs_dependencies",
)

rules_nixpkgs_dependencies()

load("//:defs.bzl", "gazelle_nix_dependencies")

gazelle_nix_dependencies()

###############
# Go preamble
###############
load(
    "@io_tweag_rules_nixpkgs//nixpkgs:toolchains/go.bzl",
    "nixpkgs_go_configure",
)

nixpkgs_go_configure(repository = "@io_tweag_gazelle_nix_nixpkgs")

load("@io_bazel_rules_go//go:deps.bzl", "go_rules_dependencies")

go_rules_dependencies()

####################
# Gazelle preamble
####################
load("@bazel_gazelle//:deps.bzl", "gazelle_dependencies")

gazelle_dependencies()
