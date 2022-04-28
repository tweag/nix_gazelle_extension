{ pkgs ? import ./nix/nixpkgs-stable.nix { } }:

with pkgs;

mkShell {
  buildInputs = [ bazel_4 binutils cacert go go-tools nix openjdk11 python3 less ];
  shellHook = ''
    mkdir -p $(pwd)/gazelle_nix/.go
    mkdir -p $(pwd)/gazelle_nix/.gocache
    ln -fs ${pkgs.go}/share/go $(pwd)/gazelle_nix/.goroot
    ln -fs ${pkgs.go-tools} $(pwd)/gazelle_nix/.gotools

    export GO11MODULE=on
    export GOCACHE=$(pwd)/gazelle_nix/.gocache
    export GOPATH=$(pwd)/gazelle_nix/.go

    # Making VSCode Go extension happy
    # (Installing latest staticcheck fails)
    go install honnef.co/go/tools/cmd/staticcheck@v0.2.2
  '';
}
