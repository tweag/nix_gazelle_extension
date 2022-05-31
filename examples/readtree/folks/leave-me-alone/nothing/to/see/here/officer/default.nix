{pkgs}:
pkgs.stdenv.mkDerivation rec {
  name = "close-one";
  src = ./src;
  buildPhase = false;
  installPhase = ''
    mkdir -p $out/bin
    cp $src/truth.source $out/bin/truth.bin
  '';
  buildInputs = [];
}
