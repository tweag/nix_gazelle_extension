{pkgs ? import ../../nixpkgs {} }:

pkgs.stdenv.mkDerivation rec {
  name = "i-need-a-friend";
  src = ./src;
  buildPhase = false;
  installPhase = ''
    mkdir $out
    cp $src/truth.source $out/truth.bin
  '';
  buildInputs = [ pkgs.mypkgs.folks.cool-kid ];
}

