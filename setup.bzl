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

def io_tweag_gazelle_nix_setup(nixpkgs = None, get_go_from_nix = True):
    """ Final initialization of gazelle nix dependencies.

    Why deps.bzl and defs.bzl are not defined together, I hear you ask?
    Splendid question! I apploud your sense of good taste.

    Bazel seems to be eagerly resolving all `load` statements, which
    results in circular dependencies happening, whenever we try to load
    both rules_nixpkgs_dependencies and other rukes_nixpkgs:*.bzl in
    the same macro.
    """
    if not nixpkgs:
        nixpkgs = "@io_tweag_gazelle_nix_nixpkgs"
        nixpkgs_local_repository(
            name = nixpkgs[1:],
            nix_file = "@io_tweag_gazelle_nix//third_party/nix:nixpkgs.nix",
            nix_file_deps = [
                "@io_tweag_gazelle_nix//third_party/nix:nixpkgs.json",
            ],
        )

    if get_go_from_nix:
        nixpkgs_go_configure(
            repository = nixpkgs,
            nix_file_content = """
            with import <nixpkgs> { config = {}; overlays = []; }; buildEnv {
              name = "bazel-go-toolchain";
              paths = [
                go_1_18
              ];
              postBuild = ''
                touch $out/ROOT
                ln -s $out/share/go/{api,doc,lib,misc,pkg,src} $out/
              '';
            }
        """,
        )

    gazelle_dependencies()

    nixpkgs_package(
        name = "fptrace",
        nix_file = "@io_tweag_gazelle_nix//third_party/nix:fptrace.nix",
        nix_file_deps = [
            "@io_tweag_gazelle_nix//third_party/nix:001-fptrace_lstat.patch",
        ],
        repository = nixpkgs,
    )
