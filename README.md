```
nix-shell
bazel run @nixscan//:bin/nixscan -- "$(pwd)/examples/folks/cool-kid/default.nix"
bazel run //language/nix:go_default_binary
bazel run //:gazelle-update
git status
```
