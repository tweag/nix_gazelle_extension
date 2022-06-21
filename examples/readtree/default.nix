{system ? builtins.currentSystem}: let
  nixpkgs = import ./nix/nixpkgs {inherit system;};

  readPkgs = import ./nix/readPkgs {
    callPackage = callPackage';
    lib = nixpkgs.lib;
  };

  callPackage' = nixpkgs.lib.callPackageWith {
    mypkgs = readPkgs ./.;
    pkgs = nixpkgs;
    lib = nixpkgs.lib;
  };
in
  nixpkgs // (readPkgs ./.)
