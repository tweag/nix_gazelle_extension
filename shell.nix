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
    ln -fs ${pkgs.go}/share/go $(pwd)/.goroot
    ln -fs ${pkgs.go-tools} $(pwd)/.gotools

    export GO11MODULE=on
    export GOCACHE=$(pwd)/.gocache
    export GOPATH=$(pwd)/.go

    # Making VSCode Go extension happy
    # (Installing latest staticcheck fails)
    go install honnef.co/go/tools/cmd/staticcheck@v0.2.2
  '';
}
