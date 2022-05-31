{
  pkgs,
  mypkgs,
}:
pkgs.stdenv.mkDerivation rec {
  name = "42";
  src = ./src;
  buildPhase = false;
  installPhase = ''
    mkdir -p $out/bin
    cp $src/truth.source $out/bin/truth.bin
  '';
  buildInputs = with mypkgs.folks; [the-one-all-know cool-kid lone-wolf];
}
