{pkgs ? import ../../nixpkgs {}}:

pkgs.stdenv.mkDerivation rec {
  name = "the-one-all-know";
  src = ./src;
  buildPhase = false;
  installPhase = ''
    mkdir $out
    cp $src/truth.source $out/truth.bin
  '';
  buildInputs = [ pkgs.mypkgs.folks.lone-wolf ];
}

