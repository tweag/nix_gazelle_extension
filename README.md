```
nix-shell
cd examples/vanilla
GAZELLE_LANGUAGES_NIX_LOG_LEVEL=debug bazel run //:gazelle-update-all
git status
```
