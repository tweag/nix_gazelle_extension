{ pkgs ? import ./third_party/nix/nixpkgs.nix { } }:

with pkgs;

mkShell {
  buildInputs = [
    bazel_5
    bazel-buildtools
    binutils
    cacert
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
    ln -fs ${pkgs.go_1_18}/share/go $(pwd)/.goroot

    export GO111MODULE=on
    export GOCACHE=$(pwd)/.gocache
    export GOPATH=$(pwd)/.go
  '';
}
