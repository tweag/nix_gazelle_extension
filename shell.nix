{ pkgs ? import ./nix/nixpkgs-stable.nix { } }:

with pkgs;

mkShell {
  buildInputs = [ bazel_4 binutils cacert go go-tools nix openjdk11 python3 less ];
  shellHook = ''
    mkdir -p $(pwd)/.go
    mkdir -p $(pwd)/.gocache
    ln -fs ${pkgs.go}/share/go $(pwd)/.goroot
    ln -fs ${pkgs.go-tools} $(pwd)/.gotools

    export GO11MODULE=on
    export GOCACHE=$(pwd).gocache
    export GOPATH=$(pwd)/.go
  '';
}
