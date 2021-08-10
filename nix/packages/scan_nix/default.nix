{ system ? builtins.currentSystem, pkgs ? import <nixpkgs> { inherit system; }
, ... }:

pkgs.rustPlatform.buildRustPackage {
  pname = "scan-nix";
  version = "0.1.0";

  src = builtins.path {
    path = ./.;
    name = "scan-nix-src";
  };

  # For updating the hash
  #cargoHash = "sha256-AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=";
  cargoHash = "sha256-o2FXqrfko4zZNvsjr/KGSSIOaXtArUABaWyNI6IP8k0=";

  BUILD_REV_COUNT = 1;
  RUN_TIME_CLOSURE = pkgs.callPackage ./runtime.nix { };
}
