{ callPackage, lib }:
# awesome readPkgs by courtesy of adisbladis
# Walk a directory structure and create corresponding nested attribute sets of derivations

let
  inherit (builtins) readDir attrNames filter substring pathExists elem foldl';
  inherit (lib) filterAttrs;

  readPkgs = { root }:
    let
      joinChild = child: root + "/${child}";
      joinChildren = children:
        foldl' (acc: child: acc + "/${child}") root children;

      # List of non-ignored subdirectories to iterate further into
      dirs = let
        files = readDir root;
        # Only iterate over non-hidden directories
        all = attrNames
          (filterAttrs (n: v: v == "directory" && substring 0 1 n != ".")
            files);
        # Check if subdirectory should be ignored
        unignored =
          filter (d: !pathExists (root + "/${d}/.nix-ignore-subdirectory")) all;
      in unignored;

      # Use callPackage on subdirectories with a default.nix
      # toCall = filter (d: pathExists (root + "/${d}/default.nix")) dirs;
      toCall = filter (d: pathExists (joinChildren [ d "default.nix" ])) dirs;
      # And iterate further into directories without one
      toRead = filter (d: !elem d toCall) dirs;

      # Create attrset with called packages
      subPkgs = (lib.listToAttrs (map (d: {
        name = d;
        value = callPackage (joinChild d) { };
      }) toCall));
      # And read further into directories
      subTrees = (lib.listToAttrs (map (d: {
        name = d;
        value = readPkgs { root = joinChild d; };
      }) toRead));

    in subPkgs // subTrees;

in root: readPkgs { inherit root; }

