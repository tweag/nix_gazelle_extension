{pkgs, mypkgs}:

pkgs.stdenv.mkDerivation rec {
  name = "the-one-all-know";
  src = ./src;
  buildPhase = false;
  installPhase = ''
    mkdir $out
    cp $src/truth.source $out/truth.bin
  '';
  buildInputs = [ mypkgs.folks.lone-wolf ];
}

