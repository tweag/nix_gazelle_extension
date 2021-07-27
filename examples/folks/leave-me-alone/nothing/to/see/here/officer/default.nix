{pkgs}:

pkgs.stdenv.mkDerivation rec {
  name = "close-one";
  src = ./src;
  buildPhase = false;
  installPhase = ''
    mkdir $out
    cp $src/truth.source $out/truth.bin
  '';
  buildInputs = [ ];
}

