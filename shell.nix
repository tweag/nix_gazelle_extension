{ pkgs ? import ./nix/nixpkgs-stable.nix {} }:

with pkgs;

mkShell {
  buildInputs = [
    bazel_4
    binutils
    cacert
    nix
    openjdk11
    python3
    less
  ];
}
