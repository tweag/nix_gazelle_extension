{ system ? builtins.currentSystem, ... }:
import (builtins.fetchTarball {
  name = "nixos-21.05-2021-08-12";
  # URL obtained from https://status.nixos.org/
  url =
    "https://github.com/NixOS/nixpkgs/archive/927ce1afc1db40869a463a37ea2738c27d425f80.tar.gz";
  # Hash obtained using `nix-prefetch-url --unpack <url>`
  sha256 = "1f64kkjk0ba9hzf086nkvk04wkfgjgzlkjpg49nrj3ar0chvzkrb";
}) {
  inherit system;
  config.allowUnfree = true;
}

