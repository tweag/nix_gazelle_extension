load("@bazel_tools//tools/build_defs/repo:http.bzl", "http_archive")
load("@bazel_tools//tools/build_defs/repo:utils.bzl", "maybe")

def io_tweag_gazelle_nix_repositories():
    """ Load repositories required by gazelle_nix extension. """
    maybe(
        http_archive,
        name = "io_tweag_rules_nixpkgs",
        sha256 = "941b0a6c45eb60b8245d79d71053d608499d1a9aba166c7a1b34d40c95112b4a",
        strip_prefix = "rules_nixpkgs-%s" % "027dc977d0cb8d93c43a5ae0e812c779e88beb9a",
        urls = ["https://github.com/tweag/rules_nixpkgs/archive/%s.tar.gz" % "027dc977d0cb8d93c43a5ae0e812c779e88beb9a"],
    )

    maybe(
        http_archive,
        name = "io_bazel_rules_go",
        sha256 = "f2dcd210c7095febe54b804bb1cd3a58fe8435a909db2ec04e31542631cf715c",
        urls = [
            "https://mirror.bazel.build/github.com/bazelbuild/rules_go/releases/download/v0.31.0/rules_go-v0.31.0.zip",
            "https://github.com/bazelbuild/rules_go/releases/download/v0.31.0/rules_go-v0.31.0.zip",
        ],
    )

    maybe(
        http_archive,
        name = "bazel_gazelle",
        sha256 = "5982e5463f171da99e3bdaeff8c0f48283a7a5f396ec5282910b9e8a49c0dd7e",
        urls = [
            "https://mirror.bazel.build/github.com/bazelbuild/bazel-gazelle/releases/download/v0.25.0/bazel-gazelle-v0.25.0.tar.gz",
            "https://github.com/bazelbuild/bazel-gazelle/releases/download/v0.25.0/bazel-gazelle-v0.25.0.tar.gz",
        ],
    )

