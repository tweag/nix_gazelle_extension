{ system ? builtins.currentSystem, ... }:
let
  nixpkgs = import ./nix/nixpkgs {
    inherit system;
  };

  readPkgs = import ./nix/readPkgs {
    callPackage = callPackage';
    lib = nixpkgs.lib; 
  };

  callPackage' = nixpkgs.lib.callPackageWith {
    mypkgs = readPkgs ./.;
    pkgs = nixpkgs;
    lib = nixpkgs.lib;
  };
  
  # Do not call self
  fileToCall = builtins.getEnv "NIX_FILE"; 
  packageToCall =
    if fileToCall == (builtins.toString ./.) + "/default.nix" then {}
    else callPackage' fileToCall {};
in packageToCall
