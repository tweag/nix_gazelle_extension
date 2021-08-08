load(
    "@io_tweag_rules_nixpkgs//nixpkgs:nixpkgs.bzl",
    "nixpkgs_local_repository",
    "nixpkgs_package",
)


def gazelle_nix_dependencies():
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
        "@io_tweag_gazelle_nix//nix:nixpkgs-stable.nix",
        "@io_tweag_gazelle_nix//nix:nixpkgs-stable.json",
        "@io_tweag_gazelle_nix//nix/packages/scan_nix:Cargo.lock",
        "@io_tweag_gazelle_nix//nix/packages/scan_nix:Cargo.toml",
        "@io_tweag_gazelle_nix//nix/packages/scan_nix:runtime-closure.nix.template",
        "@io_tweag_gazelle_nix//nix/packages/scan_nix:runtime.nix",
        "@io_tweag_gazelle_nix//nix/packages/scan_nix:default.nix",
        "@io_tweag_gazelle_nix//nix/packages/scan_nix:src/main.rs",
    ],
    repositories = {"nixpkgs": "@io_tweag_gazelle_nix_nixpkgs"},
  )
