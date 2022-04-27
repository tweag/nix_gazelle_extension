{ pkgs, mypkgs }:

pkgs.stdenv.mkDerivation rec {
  name = "the-one-all-know";
  src = ./src;
  buildPhase = false;
  installPhase = ''
    mkdir -p $out/bin
    cp $src/truth.source $out/bin/truth.bin
  '';
  buildInputs = [ mypkgs.folks.lone-wolf ];
}

