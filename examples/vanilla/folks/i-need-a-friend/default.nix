{pkgs ? import <nixpkgs> {}}:
pkgs.stdenv.mkDerivation rec {
  name = "i-need-a-friend";
  src = ./src;
  buildPhase = false;
  installPhase = ''
    mkdir -p $out/bin
    cp $src/truth.source $out/bin/truth.bin
  '';
  buildInputs = [(pkgs.callPackage ../lone-wolf {})];
}
