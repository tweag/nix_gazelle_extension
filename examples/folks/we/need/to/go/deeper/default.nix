{pkgs}:

pkgs.stdenv.mkDerivation rec {
  name = "42";
  src = ./src;
  buildPhase = false;
  installPhase = ''
    mkdir $out
    cp $src/truth.source $out/truth.bin
  '';
  buildInputs = with pkgs.mypkgs.folks; [
    the-one-all-know
    cool-kid
    lone-wolf 
  ];
}
