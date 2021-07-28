{ nixpkgs ? <nixpkgs>
, pkgs ? import nixpkgs {
    # This is a hack to work around something requiring libcap on MacOS
    config.allowUnsupportedSystem = true;
  }
, generatedCargoNix ? ./Cargo.nix
}:
let
  lorriSrc = pkgs.fetchFromGitHub {
      owner = "nix-community";
      repo = "lorri";
      rev = "dfbf9b3d22474380ee5e096931dbf25b1c162d10";
      sha256 = "0n3n0wq6s13yznljrqqsjvjxjk6lg9j4bkxpxl2l517s6012hvxs";
  };

  customBuildRustCrateForPkgs = pkgs: pkgs.buildRustCrate.override {
    defaultCrateOverrides = pkgs.defaultCrateOverrides // {
      lorri = attrs: {
        src = lorriSrc.out;
        BUILD_REV_COUNT = 1;
        RUN_TIME_CLOSURE = pkgs.callPackage "${lorriSrc.out}/nix/runtime.nix" {};
      };
      human-panic = attrs: {
        src = "${lorriSrc.out}/vendor/human-panic";
      };
    };
  };
  generatedBuild = import ./Cargo.nix {
    inherit pkgs;
    buildRustCrateForPkgs = customBuildRustCrateForPkgs;
  };
in generatedBuild.rootCrate.build
