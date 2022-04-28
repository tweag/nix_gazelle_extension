load("@bazel_tools//tools/build_defs/repo:utils.bzl", "maybe")
load("@bazel_gazelle//:deps.bzl", "gazelle_dependencies")
load(
    "@io_tweag_rules_nixpkgs//nixpkgs:nixpkgs.bzl",
    "nixpkgs_local_repository",
    "nixpkgs_package",
)
load(
    "@io_tweag_rules_nixpkgs//nixpkgs:toolchains/go.bzl",
    "nixpkgs_go_configure",
)

def io_tweag_gazelle_nix_packages(nixpkgs = None):
    """ Load packages required by gazelle_nix extension. """
    if not nixpkgs:
        nixpkgs = "@io_tweag_gazelle_nix_nixpkgs"
        nixpkgs_local_repository(
            name = nixpkgs[1:],
            nix_file = "@io_tweag_gazelle_nix//nix:nixpkgs-stable.nix",
            nix_file_deps = [
                "@io_tweag_gazelle_nix//nix:nixpkgs-stable.json",
            ],
        )

    nixpkgs_go_configure(
        repository = nixpkgs,
    )

    nixpkgs_package(
        name = "fptrace",
        nix_file = "@io_tweag_gazelle_nix//nix:fptrace.nix",
        repository = nixpkgs,
    )

    gazelle_dependencies()
