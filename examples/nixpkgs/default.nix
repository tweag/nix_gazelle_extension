{ system ? builtins.currentSystem, ... }:
let
  # Ensure used pkgs are in sync with flakes
  flakes_spec = builtins.fromJSON (builtins.readFile ./flake.lock);
  nixpkgs_spec = flakes_spec.nodes.pkgs.locked;
  nixpkgs_src = builtins.fetchTarball {
    url =
      "https://github.com/${nixpkgs_spec.owner}/${nixpkgs_spec.repo}/archive/${nixpkgs_spec.rev}.tar.gz";
    sha256 =
      (builtins.replaceStrings [ "sha256-" ] [ "" ] nixpkgs_spec.narHash);
  };
  nixpkgs = import nixpkgs_src {
    inherit system;
    config.allowUnfree = true;
    overlays = [ (self: super: {
      mypkgs = ((super.callPackage ../toys { }) ./..);
    }
    )];
  };

in nixpkgs

