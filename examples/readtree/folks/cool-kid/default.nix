{
  pkgs,
  mypkgs,
}:
pkgs.stdenv.mkDerivation rec {
  name = "cool-kid";
  src = ./src;
  buildPhase = false;
  installPhase = ''
    mkdir -p $out/bin
    cp $src/truth.source $out/bin/truth.bin
  '';
  buildInputs = [mypkgs.folks.the-one-all-know];
}
