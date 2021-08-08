workspace(
    name = "io_tweag_gazelle_nix",
)

load(
    "@bazel_tools//tools/build_defs/repo:http.bzl",
    "http_archive",
)

RULES_TWEAG_COMMIT = "a388ab60dea07c3fc182453e89ff1a67c9d3eba6"

http_archive(
    name = "io_tweag_rules_nixpkgs",
    sha256 = "6bedf80d6cb82d3f1876e27f2ff9a2cc814d65f924deba14b49698bb1fb2a7f7",
    strip_prefix = "rules_nixpkgs-%s" % RULES_TWEAG_COMMIT,
    urls = ["https://github.com/tweag/rules_nixpkgs/archive/%s.tar.gz" % RULES_TWEAG_COMMIT],
)

load(
    "@io_tweag_rules_nixpkgs//nixpkgs:repositories.bzl",
    "rules_nixpkgs_dependencies",
)

rules_nixpkgs_dependencies()

load(
    "@io_tweag_rules_nixpkgs//nixpkgs:nixpkgs.bzl",
    "nixpkgs_local_repository",
    "nixpkgs_package",
)

nixpkgs_local_repository(
    name = "nixpkgs",
    nix_file = "//nix:nixpkgs-stable.nix",
    nix_file_deps = [
        "//nix:nixpkgs-stable.json",
    ],
)

nixpkgs_package(
    name = "scan_nix",
    nix_file = "//nix/packages/scan_nix:default.nix",
    nix_file_deps = [
        "//nix:nixpkgs-stable.nix",
        "//nix:nixpkgs-stable.json",
        "//nix/packages/scan_nix:Cargo.lock",
        "//nix/packages/scan_nix:Cargo.toml",
        "//nix/packages/scan_nix:runtime-closure.nix.template",
        "//nix/packages/scan_nix:runtime.nix",
        "//nix/packages/scan_nix:default.nix",
        "//nix/packages/scan_nix:src/main.rs",
    ],
    repositories = {"nixpkgs": "@nixpkgs"},
)

http_archive(
    name = "io_bazel_rules_go",
    sha256 = "69de5c704a05ff37862f7e0f5534d4f479418afc21806c887db544a316f3cb6b",
    urls = [
        "https://mirror.bazel.build/github.com/bazelbuild/rules_go/releases/download/v0.27.0/rules_go-v0.27.0.tar.gz",
        "https://github.com/bazelbuild/rules_go/releases/download/v0.27.0/rules_go-v0.27.0.tar.gz",
    ],
)

http_archive(
    name = "bazel_gazelle",
    sha256 = "62ca106be173579c0a167deb23358fdfe71ffa1e4cfdddf5582af26520f1c66f",
    urls = [
        "https://mirror.bazel.build/github.com/bazelbuild/bazel-gazelle/releases/download/v0.23.0/bazel-gazelle-v0.23.0.tar.gz",
        "https://github.com/bazelbuild/bazel-gazelle/releases/download/v0.23.0/bazel-gazelle-v0.23.0.tar.gz",
    ],
)

load(
    "@io_tweag_rules_nixpkgs//nixpkgs:toolchains/go.bzl",
    "nixpkgs_go_configure",
)

nixpkgs_go_configure(repositories = {"nixpkgs": "@nixpkgs"})

load("@io_bazel_rules_go//go:deps.bzl", "go_rules_dependencies")
load("@bazel_gazelle//:deps.bzl", "gazelle_dependencies")

go_rules_dependencies()

gazelle_dependencies()
