{ pkgs ? import ./third_party/nix/nixpkgs.nix { } }:

with pkgs;

mkShell {
  buildInputs = [
    bazel_5
    binutils
    cacert
    git
    go_1_18
    go-tools
    nix
    nixfmt
    openjdk11
    python3
    less
  ];
  shellHook = ''
    mkdir -p $(pwd)/.go
    mkdir -p $(pwd)/.gocache

    export GO111MODULE=on
    export GOCACHE=$(pwd)/.gocache
    export GOENV=$(pwd)/env
    export GOPATH=$(pwd)/.go
  '';
}
