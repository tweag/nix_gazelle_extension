This project is a gazelle extension that generates and updates [rules_nixpkgs](https://github.com/tweag/rules_nixpkgs) definitions for Bazel from your workspace.

For each directory containing a `default.nix` file, an appropriate `nixpkgs_package` external repository is created within the `WORKSPACE` file, as well as any supporting definitions needed to make derivation work when invoked from the Bazel. Any required files or dependent nix derivations are traced and captured as long as they are part of your workspace.

To see the extension in action:

- `$ bazel run //examples:generate`
- `$ cd examples/vanilla`
- `$ bazel run //:gazelle-update-all`

You can learn more about the extension setup by browsing the `examples` subdirectory.
