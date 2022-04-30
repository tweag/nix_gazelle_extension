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
