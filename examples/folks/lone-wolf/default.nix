{pkgs ? import ../../nixpkgs {}}:

pkgs.stdenv.mkDerivation rec {
  name = "lone-wolf";
  src = ./src;
  buildPhase = false;
  installPhase = ''
    mkdir $out
    cp $src/truth.source $out/truth.bin
  '';
  buildInputs = [ ];
}

