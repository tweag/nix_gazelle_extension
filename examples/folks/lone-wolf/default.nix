{pkgs}:

pkgs.stdenv.mkDerivation rec {
  name = "lone-wolf";
  src = ./src;
  buildPhase = false;
  installPhase = ''
    mkdir -p $out/bin
    cp $src/truth.source $out/bin/truth.bin
  '';
  buildInputs = [ ];
}

