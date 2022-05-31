{pkgs ? import ./third_party/nix/nixpkgs.nix {}}:
with pkgs; let
  inputs = [
    # formatters
    treefmt
    alejandra
    bazel-buildtools
    shellcheck
    shfmt
    nodePackages.prettier
    nodePackages.prettier-plugin-toml
    gotools
    go-tools

    bazel_5

    direnv
    binutils
    cacert
    git
    go_1_18
    nix
    openjdk11
    python3
    less
  ];
in
  mkShell {
    buildInputs = inputs;
    shellHook = ''
      mkdir -p $(pwd)/.go
      mkdir -p $(pwd)/.gocache

      export GO111MODULE=on
      export GOCACHE=$(pwd)/.gocache
      export GOENV=$(pwd)/env
      export GOPATH=$(pwd)/.go

      export NODE_PATH=${pkgs.nodePackages.prettier-plugin-toml}/lib/node_modules:$NODE_PATH
    '';
  }
