{ system ? builtins.currentSystem, nix_file ? null, }:
let
  nixpkgs = import ./nix/nixpkgs { inherit system; };

  readPkgs = import ./nix/readPkgs {
    callPackage = callPackage';
    lib = nixpkgs.lib;
  };

  callPackage' = nixpkgs.lib.callPackageWith {
    mypkgs = readPkgs ./.;
    pkgs = nixpkgs;
    lib = nixpkgs.lib;
  };

  _nix_file =
    if builtins.isNull nix_file then builtins.getEnv "NIX_FILE" else nix_file;

  _nix_file' = if builtins.stringLength _nix_file == 0 then
    abort "File to evaluate missing."
  else
    _nix_file;

  fileToCall = if nixpkgs.lib.hasPrefix "/" _nix_file' then
    _nix_file'
  else
    "${(builtins.toString ./.)}/${_nix_file'}";

  # Do not call self
  packageToCall = if fileToCall == (builtins.toString ./.) + "/default.nix" then
    { }
  else
    callPackage' fileToCall { };
in packageToCall

