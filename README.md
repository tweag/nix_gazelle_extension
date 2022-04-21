```
nix-shell
cd examples/readtree
bazel run //:gazelle-update
bazel run //:gazelle-update-repos
git status
```
