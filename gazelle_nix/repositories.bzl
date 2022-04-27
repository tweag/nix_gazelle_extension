load("@bazel_tools//tools/build_defs/repo:http.bzl", "http_archive")
load("@bazel_tools//tools/build_defs/repo:utils.bzl", "maybe")

def io_tweag_gazelle_nix_repositories():
    """ Load repositories required by gazelle_nix extension. """
    maybe(
        http_archive,
        name = "io_tweag_rules_nixpkgs",
        sha256 = "6bedf80d6cb82d3f1876e27f2ff9a2cc814d65f924deba14b49698bb1fb2a7f7",
        strip_prefix = "rules_nixpkgs-%s" % "a388ab60dea07c3fc182453e89ff1a67c9d3eba6",
        urls = ["https://github.com/tweag/rules_nixpkgs/archive/%s.tar.gz" % "a388ab60dea07c3fc182453e89ff1a67c9d3eba6"],
    )

    maybe(
        http_archive,
        name = "io_bazel_rules_go",
        sha256 = "69de5c704a05ff37862f7e0f5534d4f479418afc21806c887db544a316f3cb6b",
        urls = [
            "https://mirror.bazel.build/github.com/bazelbuild/rules_go/releases/download/v0.27.0/rules_go-v0.27.0.tar.gz",
            "https://github.com/bazelbuild/rules_go/releases/download/v0.27.0/rules_go-v0.27.0.tar.gz",
        ],
    )

    maybe(
        http_archive,
        name = "bazel_gazelle",
        sha256 = "62ca106be173579c0a167deb23358fdfe71ffa1e4cfdddf5582af26520f1c66f",
        urls = [
            "https://mirror.bazel.build/github.com/bazelbuild/bazel-gazelle/releases/download/v0.23.0/bazel-gazelle-v0.23.0.tar.gz",
            "https://github.com/bazelbuild/bazel-gazelle/releases/download/v0.23.0/bazel-gazelle-v0.23.0.tar.gz",
        ],
    )

