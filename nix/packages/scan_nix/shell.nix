{ system ? builtins.currentSystem, pkgs ? import <nixpkgs> { inherit system; }
, ... }:

pkgs.mkShell {
  name = "scan-nix-dev-shell";
  nativeBuildInputs = with pkgs; [ cargo rustfmt ];

  # Lorri requirements
  BUILD_REV_COUNT = 1;
  RUN_TIME_CLOSURE = pkgs.callPackage ./runtime.nix { };

  shellHook = ''
    export TERM=xterm
    echo 'Welcome to scan-nix development shell!'
  '';
}
