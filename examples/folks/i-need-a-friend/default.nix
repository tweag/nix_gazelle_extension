{pkgs, mypkgs}:

pkgs.stdenv.mkDerivation rec {
  name = "i-need-a-friend";
  src = ./src;
  buildPhase = false;
  installPhase = ''
    mkdir $out
    cp $src/truth.source $out/truth.bin
  '';
  buildInputs = [ mypkgs.folks.cool-kid ];
}

