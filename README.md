```
nix-shell
cd examples/vanilla
GAZELLE_LANGUAGES_NIX_LOG_LEVEL=debug bazel run //:gazelle-update
GAZELLE_LANGUAGES_NIX_LOG_LEVEL=debug bazel run //:gazelle-update-repos
git status
```
