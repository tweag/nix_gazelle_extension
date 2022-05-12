workspace(
    name = "io_tweag_gazelle_nix",
)

load("//:repositories.bzl", "io_tweag_gazelle_nix_repositories")

io_tweag_gazelle_nix_repositories()

load("//:deps.bzl", "io_tweag_gazelle_nix_deps")

# gazelle:repository_macro third_party/go/repositories.bzl%go_deps
io_tweag_gazelle_nix_deps()

load("//:setup.bzl", "io_tweag_gazelle_nix_setup")

io_tweag_gazelle_nix_setup()

load("@bazel_tools//tools/build_defs/repo:http.bzl", "http_archive")
load("@bazel_tools//tools/build_defs/repo:utils.bzl", "maybe")

maybe(
    http_archive,
    name = "rules_proto",
    sha256 = "66bfdf8782796239d3875d37e7de19b1d94301e8972b3cbd2446b332429b4df1",
    strip_prefix = "rules_proto-4.0.0",
    url = "https://github.com/bazelbuild/rules_proto/archive/refs/tags/4.0.0.tar.gz",
)

maybe(
    http_archive,
    name = "com_google_protobuf",
    sha256 = "c6003e1d2e7fefa78a3039f19f383b4f3a61e81be8c19356f85b6461998ad3db",
    strip_prefix = "protobuf-3.17.3",
    urls = ["https://github.com/protocolbuffers/protobuf/archive/v3.17.3.tar.gz"],
)

load("@rules_proto//proto:repositories.bzl", "rules_proto_dependencies")

rules_proto_dependencies()
