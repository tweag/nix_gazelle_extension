load("@bazel_gazelle//:deps.bzl", "gazelle_dependencies")
load("@io_bazel_rules_go//go:deps.bzl", "go_rules_dependencies")
load(
    "@io_tweag_rules_nixpkgs//nixpkgs:nixpkgs.bzl",
    "nixpkgs_local_repository",
    "nixpkgs_package",
)
load(
   "@io_tweag_rules_nixpkgs//nixpkgs:repositories.bzl",
   "rules_nixpkgs_dependencies",
)
load(
    "@io_tweag_rules_nixpkgs//nixpkgs:toolchains/go.bzl",
    "nixpkgs_go_configure",
)

def io_tweag_gazelle_nix_packages(nixpkgs = "@io_tweag_gazelle_nix_nixpkgs"):
    rules_nixpkgs_dependencies()

    nixpkgs_local_repository(
        name = "io_tweag_gazelle_nix_nixpkgs",
        nix_file = "@io_tweag_gazelle_nix//nix:nixpkgs-stable.nix",
        nix_file_deps = [
            "@io_tweag_gazelle_nix//nix:nixpkgs-stable.json",
        ],
    )

    nixpkgs_package(
        name = "scan_nix",
        nix_file = "@io_tweag_gazelle_nix//nix/packages/scan_nix:default.nix",
        nix_file_deps = [
            "@io_tweag_gazelle_nix//nix/packages/scan_nix:Cargo.lock",
            "@io_tweag_gazelle_nix//nix/packages/scan_nix:Cargo.toml",
            "@io_tweag_gazelle_nix//nix/packages/scan_nix:runtime-closure.nix.template",
            "@io_tweag_gazelle_nix//nix/packages/scan_nix:runtime.nix",
            "@io_tweag_gazelle_nix//nix/packages/scan_nix:default.nix",
            "@io_tweag_gazelle_nix//nix/packages/scan_nix:src/main.rs",
        ],
        repository = nixpkgs,
    )
    
    nixpkgs_go_configure(repository = nixpkgs)
    go_rules_dependencies()
    gazelle_dependencies()
