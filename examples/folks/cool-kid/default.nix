{pkgs, mypkgs}:

pkgs.stdenv.mkDerivation rec {
  name = "cool-kid";
  src = ./src;
  buildPhase = false;
  installPhase = ''
    mkdir $out
    cp $src/truth.source $out/truth.bin
  '';
  buildInputs = [ mypkgs.folks.the-one-all-know ];
}

