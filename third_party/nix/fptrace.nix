{pkgs ? import <nixpkgs> {}}:
import
(
  pkgs.applyPatches {
    name = "fptrace";
    src = fetchTarball {
      url = "https://github.com/orivej/fptrace/archive/6d1a0ee777e2b3441a615f8bfc2249833292d920.tar.gz";
      sha256 = "1v9mqnh9m9l2zdjp59yczssdx3v1qmcz8y815ajnigzdbny2b5vk";
    };
    patches = [./001-fptrace_lstat.patch];
  }
)
